package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/users"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailTaken          = errors.New("email already taken")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrInvalidTOTPCode     = errors.New("invalid TOTP code")
	ErrInvalidRecoveryCode = errors.New("invalid recovery code")
	ErrRecoveryCodeUsed    = errors.New("recovery code already used")
	ErrTOTPNotEnabled      = errors.New("TOTP not enabled")
	ErrPasskeyNotEnabled   = errors.New("passkey not enabled")
	ErrSessionNotFound     = errors.New("session not found")
	ErrWrongPassword       = errors.New("wrong password")
)

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=100"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required"`
	TOTPCode *string `json:"totp_code"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type TOTPVerifyRequest struct {
	Code string `json:"code" binding:"required"`
}

type TOTPDisableRequest struct {
	Password string `json:"password" binding:"required"`
}

type PasskeyRegisterRequest struct {
	Name string `json:"name"`
}

type PasskeyLoginRequest struct {
	Credential string `json:"credential"`
}

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type passwordResetClaims struct {
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type emailVerifyClaims struct {
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db       *gorm.DB
	cfg      *configs.AuthConfig
	webAuthn *webauthn.WebAuthn
}

func NewService(db *gorm.DB, cfg *configs.AuthConfig) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		s.cfg.Argon2Time,
		s.cfg.Argon2Memory,
		s.cfg.Argon2Threads,
		s.cfg.Argon2KeyLength,
	)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		s.cfg.Argon2Memory,
		s.cfg.Argon2Time,
		s.cfg.Argon2Threads,
		saltB64,
		hashB64,
	)

	return encoded, nil
}

func (s *AuthService) VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}
	if version != 19 {
		return false, errors.New("unsupported argon2 version")
	}

	var memory uint32
	var timeVal uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeVal, &threads); err != nil {
		return false, fmt.Errorf("invalid parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %w", err)
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		timeVal,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	if len(computedHash) != len(expectedHash) {
		return false, nil
	}

	for i := range computedHash {
		if computedHash[i] != expectedHash[i] {
			return false, nil
		}
	}

	return true, nil
}

func (s *AuthService) GenerateTokenPair(user *users.User, ip, userAgent string) (accessToken, refreshToken string, err error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := &Claims{
		Email: user.Email,
		Role:  string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTExpiration)),
			Issuer:    "life-log",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken = hex.EncodeToString(refreshBytes)

	refreshHash := sha256.Sum256([]byte(refreshToken))
	tokenHash := sha256.Sum256([]byte(jti))

	session := &users.Session{
		UserID:       user.ID,
		Token:        hex.EncodeToString(tokenHash[:]),
		RefreshToken: hex.EncodeToString(refreshHash[:]),
		IPAddress:    ip,
		UserAgent:    userAgent,
		ExpiresAt:    now.Add(s.cfg.RefreshTokenExpiry),
	}

	if err := s.db.Create(session).Error; err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	refreshHash := sha256.Sum256([]byte(tokenString))
	hashHex := hex.EncodeToString(refreshHash[:])

	var session users.Session
	if err := s.db.Where("refresh_token = ? AND is_revoked = ? AND expires_at > ?",
		hashHex, false, time.Now()).
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, ErrInvalidToken
		}
		return uuid.Nil, fmt.Errorf("failed to query session: %w", err)
	}

	if err := s.db.Model(&session).Update("is_revoked", true).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to revoke old session: %w", err)
	}

	return session.UserID, nil
}

func (s *AuthService) RevokeSession(tokenString string) error {
	refreshHash := sha256.Sum256([]byte(tokenString))
	hashHex := hex.EncodeToString(refreshHash[:])

	result := s.db.Model(&users.Session{}).
		Where("refresh_token = ?", hashHex).
		Update("is_revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *AuthService) RevokeAllUserSessions(userID uuid.UUID) error {
	result := s.db.Model(&users.Session{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Update("is_revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", result.Error)
	}

	return nil
}

func (s *AuthService) GenerateTOTPSecret(userEmail string) (secret string, image string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.TOTPIssuer,
		AccountName: userEmail,
		Period:      30,
		Digits:      6,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	image = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	secret = key.Secret()

	return secret, image, nil
}

func (s *AuthService) ValidateTOTPCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (s *AuthService) GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, s.cfg.RecoveryCodesCount)
	for i := 0; i < s.cfg.RecoveryCodesCount; i++ {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}
		codes[i] = hex.EncodeToString(b)
	}
	return codes, nil
}

func (s *AuthService) ValidateRecoveryCode(userID uuid.UUID, code string) (bool, error) {
	var recoveryCode RecoveryCode
	if err := s.db.Where("user_id = ? AND code = ? AND is_used = ?",
		userID, code, false).
		First(&recoveryCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrInvalidRecoveryCode
		}
		return false, fmt.Errorf("failed to query recovery code: %w", err)
	}

	now := time.Now()
	if err := s.db.Model(&recoveryCode).Updates(map[string]interface{}{
		"is_used": true,
		"used_at": &now,
	}).Error; err != nil {
		return false, fmt.Errorf("failed to mark recovery code as used: %w", err)
	}

	return true, nil
}

func (s *AuthService) LogAudit(userID *uuid.UUID, action AuditAction, ip, userAgent, details string) {
	entry := &AuditLog{
		UserID:    userID,
		Action:    action,
		IPAddress: ip,
		UserAgent: userAgent,
		Details:   details,
	}

	if err := s.db.Create(entry).Error; err != nil {
		fmt.Printf("failed to log audit: %v\n", err)
	}
}

func (s *AuthService) GeneratePasswordResetToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := &passwordResetClaims{
		Purpose: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.PasswordResetExpiry)),
			Issuer:    "life-log",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign password reset token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) ValidatePasswordResetToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &passwordResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*passwordResetClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	if claims.Purpose != "password_reset" {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func (s *AuthService) generateEmailVerifyToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := &emailVerifyClaims{
		Purpose: "email_verify",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.EmailVerifyExpiry)),
			Issuer:    "life-log",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign email verify token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) validateEmailVerifyToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &emailVerifyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*emailVerifyClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	if claims.Purpose != "email_verify" {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func getClientIP(c *gin.Context) string {
	if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}
	return c.ClientIP()
}

func (s *AuthService) getUserFromContext(c *gin.Context) (*users.User, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, ErrInvalidToken
	}

	var user users.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

func (s *AuthService) HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser users.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": ErrEmailTaken.Error()})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": ErrUsernameTaken.Error()})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &users.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		DisplayName:  req.DisplayName,
		Role:         users.RoleUser,
	}

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	if err := s.db.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	accessToken, refreshToken, err := s.GenerateTokenPair(user, ip, ua)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	verifyToken, err := s.generateEmailVerifyToken(user.ID)
	if err == nil {
		_ = verifyToken
	}

	details, _ := json.Marshal(map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})
	s.LogAudit(&user.ID, ActionRegister, ip, ua, string(details))

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         user.Role,
		},
	})
}

func (s *AuthService) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user users.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidCredentials.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	valid, err := s.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidCredentials.Error()})
		return
	}

	if user.TOTPEnabled {
		if req.TOTPCode == nil || *req.TOTPCode == "" {
			c.JSON(http.StatusOK, gin.H{"requires_totp": true})
			return
		}
		if !s.ValidateTOTPCode(user.TOTPSecret, *req.TOTPCode) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidTOTPCode.Error()})
			return
		}
	}

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	accessToken, refreshToken, err := s.GenerateTokenPair(&user, ip, ua)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_ip":       ip,
		"last_user_agent": ua,
	})

	details, _ := json.Marshal(map[string]interface{}{
		"email": req.Email,
	})
	s.LogAudit(&user.ID, ActionLogin, ip, ua, string(details))

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         user.Role,
		},
	})
}

func (s *AuthService) HandleLogout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.RevokeSession(req.RefreshToken); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "logged out"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke session"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("userID"))
	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	s.LogAudit(&userID, ActionLogout, ip, ua, "")

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (s *AuthService) HandleRefresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := s.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var user users.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUserNotFound.Error()})
		return
	}

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	accessToken, refreshToken, err := s.GenerateTokenPair(&user, ip, ua)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (s *AuthService) HandleVerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := s.validateEmailVerifyToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	if err := s.db.Model(&users.User{}).Where("id = ?", userID).
		Update("email_verified_at", &now).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	s.LogAudit(&userID, ActionEmailVerify, ip, ua, "")

	c.JSON(http.StatusOK, gin.H{"message": "email verified"})
}

func (s *AuthService) HandleForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user users.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	token, err := s.GeneratePasswordResetToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "if the email exists, a reset link has been sent",
		"token":   token,
	})
}

func (s *AuthService) HandleResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := s.ValidatePasswordResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := s.db.Model(&users.User{}).Where("id = ?", userID).
		Update("password_hash", passwordHash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	if err := s.RevokeAllUserSessions(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke sessions"})
		return
	}

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	s.LogAudit(&userID, ActionPasswordReset, ip, ua, "")

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}

func (s *AuthService) HandleTOTPSetup(c *gin.Context) {
	user, err := s.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.TOTPEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP already enabled"})
		return
	}

	secret, image, err := s.GenerateTOTPSecret(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate TOTP secret"})
		return
	}

	if err := s.db.Model(user).Update("totp_secret", secret).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save TOTP secret"})
		return
	}

	recoveryCodes, err := s.GenerateRecoveryCodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate recovery codes"})
		return
	}

	for _, code := range recoveryCodes {
		rc := &RecoveryCode{
			UserID: user.ID,
			Code:   code,
		}
		if err := s.db.Create(rc).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save recovery code"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":         secret,
		"qr_code":        image,
		"recovery_codes": recoveryCodes,
	})
}

func (s *AuthService) HandleTOTPVerify(c *gin.Context) {
	user, err := s.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.TOTPSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP not set up"})
		return
	}

	if !s.ValidateTOTPCode(user.TOTPSecret, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTOTPCode.Error()})
		return
	}

	now := time.Now()
	s.db.Model(user).Updates(map[string]interface{}{
		"totp_enabled":    true,
		"totp_verified_at": &now,
	})

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	s.LogAudit(&user.ID, ActionTOTPEnable, ip, ua, "")

	c.JSON(http.StatusOK, gin.H{"message": "TOTP enabled"})
}

func (s *AuthService) HandleTOTPDisable(c *gin.Context) {
	user, err := s.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req TOTPDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := s.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrWrongPassword.Error()})
		return
	}

	s.db.Model(user).Updates(map[string]interface{}{
		"totp_enabled":    false,
		"totp_secret":     "",
		"totp_verified_at": nil,
	})

	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	s.LogAudit(&user.ID, ActionTOTPDisable, ip, ua, "")

	c.JSON(http.StatusOK, gin.H{"message": "TOTP disabled"})
}

func (s *AuthService) HandlePasskeyRegister(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WebAuthn registration not yet implemented",
	})
}

func (s *AuthService) HandlePasskeyLogin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WebAuthn login not yet implemented",
	})
}

func (s *AuthService) HandleListSessions(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
		return
	}

	var sessions []users.Session
	if err := s.db.Where("user_id = ? AND is_revoked = ? AND expires_at > ?",
		userID, false, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query sessions"})
		return
	}

	type sessionResponse struct {
		ID        uuid.UUID `json:"id"`
		IPAddress string    `json:"ip_address"`
		UserAgent string    `json:"user_agent"`
		CreatedAt time.Time `json:"created_at"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	response := make([]sessionResponse, len(sessions))
	for i, s := range sessions {
		response[i] = sessionResponse{
			ID:        s.ID,
			IPAddress: s.IPAddress,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"sessions": response})
}

func (s *AuthService) HandleRevokeSession(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	result := s.db.Model(&users.Session{}).
		Where("id = ? AND user_id = ? AND is_revoked = ?", sessionID, userID, false).
		Update("is_revoked", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke session"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrSessionNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session revoked"})
}

func (s *AuthService) HandleRevokeAllSessions(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidToken.Error()})
		return
	}

	if err := s.RevokeAllUserSessions(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all sessions revoked"})
}

func (s *AuthService) HandleAuditLogs(c *gin.Context) {
	role, exists := c.Get("userRole")
	if !exists || role.(string) != string(RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var logs []AuditLog
	query := s.db.Model(&AuditLog{}).Order("created_at DESC")

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		query = query.Where("user_id = ?", userIDStr)
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	page := 1
	pageSize := 50

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs, "page": page, "page_size": pageSize})
}
