package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/auth"
	"github.com/kunal/life-log/backend/internal/database"
	"github.com/kunal/life-log/backend/internal/users"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	authSvc    *auth.AuthService
	userSvc    *users.UserService
	testRouter *gin.Engine
)

var testDBCounter int64

func setupTest() {
	testDBCounter++
	cfg := &configs.DatabaseConfig{
		Driver: "sqlite",
		DSN:    fmt.Sprintf("file:test_%d?mode=memory&cache=private", testDBCounter),
	}
	testDB = database.Init(cfg)
	database.Migrate()

	authCfg := &configs.AuthConfig{
		JWTSecret:          "test-secret-key-not-for-production",
		JWTExpiration:      15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Argon2Time:         1,
		Argon2Memory:       64 * 1024,
		Argon2Threads:      4,
		Argon2KeyLength:    32,
		TOTPIssuer:         "LifeLogTest",
		RecoveryCodesCount: 10,
	}

	authSvc = auth.NewService(testDB, authCfg)
	userSvc = users.NewService(testDB)
	testRouter = gin.New()
}

func TestRegister(t *testing.T) {
	setupTest()

	api := testRouter.Group("/api/v1")
	mw := func(c *gin.Context) { c.Next() }
	auth.RegisterRoutes(api, authSvc, mw)

	body := `{"email":"test@example.com","username":"testuser","password":"SecurePass123!","display_name":"Test User"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Errorf("expected 200/201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	userResp, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected user in response, got %v", resp)
	}
	if userResp["email"] != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", userResp["email"])
	}
}

func TestLogin(t *testing.T) {
	setupTest()

	api := testRouter.Group("/api/v1")
	mw := func(c *gin.Context) { c.Next() }
	auth.RegisterRoutes(api, authSvc, mw)

	registerBody := `{"email":"login@example.com","username":"loginuser","password":"SecurePass123!","display_name":"Login User"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	loginBody := `{"email":"login@example.com","password":"SecurePass123!"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
	if resp["refresh_token"] == "" {
		t.Error("expected refresh_token in response")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	setupTest()

	api := testRouter.Group("/api/v1")
	mw := func(c *gin.Context) { c.Next() }
	auth.RegisterRoutes(api, authSvc, mw)

	registerBody := `{"email":"invalid@example.com","username":"invaliduser","password":"SecurePass123!","display_name":"Invalid User"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	loginBody := `{"email":"invalid@example.com","password":"WrongPassword!"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPasswordHashing(t *testing.T) {
	setupTest()

	hash, err := authSvc.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("expected argon2id hash prefix, got %s", hash)
	}

	valid, err := authSvc.VerifyPassword("testpassword123", hash)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}
	if !valid {
		t.Error("expected password verification to succeed")
	}

	invalid, err := authSvc.VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("failed to verify wrong password: %v", err)
	}
	if invalid {
		t.Error("expected password verification to fail for wrong password")
	}
}

func TestTokenGeneration(t *testing.T) {
	setupTest()

	user := &users.User{
		Email:    "token@example.com",
		Username: "tokenuser",
	}

	if err := userSvc.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	accessToken, refreshToken, err := authSvc.GenerateTokenPair(user, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	claims, err := authSvc.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.Email != "token@example.com" {
		t.Errorf("expected email token@example.com, got %s", claims.Email)
	}
}

func TestRefreshToken(t *testing.T) {
	setupTest()

	user := &users.User{
		Email:    "refresh@example.com",
		Username: "refreshuser",
	}

	if err := userSvc.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, refreshToken, err := authSvc.GenerateTokenPair(user, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	userID, err := authSvc.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if userID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, userID)
	}
}

func TestTOTPSetup(t *testing.T) {
	setupTest()

	secret, image, err := authSvc.GenerateTOTPSecret("totp@example.com")
	if err != nil {
		t.Fatalf("failed to generate TOTP secret: %v", err)
	}

	if secret == "" {
		t.Error("expected non-empty secret")
	}
	if !strings.HasPrefix(image, "data:image/png;base64,") {
		t.Error("expected base64 PNG data URI")
	}

	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	valid := authSvc.ValidateTOTPCode(secret, code)
	if !valid {
		t.Error("expected TOTP code validation to succeed")
	}
}

func TestRecoveryCodes(t *testing.T) {
	setupTest()

	user := &users.User{
		Email:    "recovery@example.com",
		Username: "recoveryuser",
	}

	if err := userSvc.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	codes, err := authSvc.GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("failed to generate recovery codes: %v", err)
	}

	if len(codes) != 10 {
		t.Errorf("expected 10 codes, got %d", len(codes))
	}

	for _, code := range codes {
		if len(code) == 0 {
			t.Error("expected non-empty recovery code")
		}
	}
}

func TestAuditLogging(t *testing.T) {
	setupTest()

	authSvc.LogAudit(nil, "test_action", "127.0.0.1", "test-agent", `{"test": true}`)

	user := &users.User{
		Email:    "audit@example.com",
		Username: "audituser",
	}
	if err := userSvc.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	authSvc.LogAudit(&user.ID, "user_action", "192.168.1.1", "chrome", `{"detail": "test"}`)

	var count int64
	testDB.Model(&auth.AuditLog{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 audit logs, got %d", count)
	}
}
