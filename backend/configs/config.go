package configs

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	MQTT     MQTTConfig
	Ntfy     NtfyConfig
	Email    EmailConfig
	CORS     CORSConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	BaseURL      string
}

type DatabaseConfig struct {
	Driver   string
	DSN      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret            string
	JWTExpiration        time.Duration
	RefreshTokenExpiry   time.Duration
	Argon2Time           uint32
	Argon2Memory         uint32
	Argon2Threads        uint8
	Argon2KeyLength      uint32
	TOTPIssuer           string
	WebAuthnRPID         string
	WebAuthnRPOrigin     string
	RecoveryCodesCount   int
	PasswordResetExpiry  time.Duration
	EmailVerifyExpiry    time.Duration
	BcryptBreachCheckURL string
}

type MQTTConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
	QoS      byte
}

type NtfyConfig struct {
	URL       string
	Topic     string
	Token     string
	Enabled   bool
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type RateLimitConfig struct {
	Enabled     bool
	RequestsPerMin int
	BurstSize   int
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			DSN:      getEnv("DB_DSN", ""),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "lifelog"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "lifelog"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:            getEnv("JWT_SECRET", "change-me-in-production"),
			JWTExpiration:        getDuration("JWT_EXPIRATION", 15*time.Minute),
			RefreshTokenExpiry:   getDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			Argon2Time:           getUint32("ARGON2_TIME", 3),
			Argon2Memory:         getUint32("ARGON2_MEMORY", 64*1024),
			Argon2Threads:        getUint8("ARGON2_THREADS", 4),
			Argon2KeyLength:      getUint32("ARGON2_KEY_LENGTH", 32),
			TOTPIssuer:           getEnv("TOTP_ISSUER", "LifeLog"),
			WebAuthnRPID:         getEnv("WEBAUTHN_RPID", "localhost"),
			WebAuthnRPOrigin:     getEnv("WEBAUTHN_RP_ORIGIN", "http://localhost:5173"),
			RecoveryCodesCount:   getInt("RECOVERY_CODES_COUNT", 10),
			PasswordResetExpiry:  getDuration("PASSWORD_RESET_EXPIRY", 1*time.Hour),
			EmailVerifyExpiry:    getDuration("EMAIL_VERIFY_EXPIRY", 24*time.Hour),
			BcryptBreachCheckURL: getEnv("BREACH_CHECK_URL", "https://haveibeenpwned.com/api/v3/pwnedpassword/"),
		},
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
			ClientID: getEnv("MQTT_CLIENT_ID", "life-log-backend"),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
			QoS:      byte(getInt("MQTT_QOS", 1)),
		},
		Ntfy: NtfyConfig{
			URL:     getEnv("NTFY_URL", "https://ntfy.sh"),
			Topic:   getEnv("NTFY_TOPIC", "life-log"),
			Token:   getEnv("NTFY_TOKEN", ""),
			Enabled: getEnv("NTFY_ENABLED", "false") == "true",
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromAddress:  getEnv("SMTP_FROM_ADDRESS", "noreply@lifelog.app"),
			FromName:     getEnv("SMTP_FROM_NAME", "LifeLog"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvSlice("CORS_ORIGINS", []string{"http://localhost:5173"}),
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
		RateLimit: RateLimitConfig{
			Enabled:        getEnv("RATE_LIMIT_ENABLED", "true") == "true",
			RequestsPerMin: getInt("RATE_LIMIT_REQUESTS", 60),
			BurstSize:      getInt("RATE_LIMIT_BURST", 10),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getUint32(key string, fallback uint32) uint32 {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			return uint32(i)
		}
	}
	return fallback
}

func getUint8(key string, fallback uint8) uint8 {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.ParseUint(v, 10, 8)
		if err == nil {
			return uint8(i)
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := splitAndTrim(v, ",")
		if len(parts) > 0 {
			return parts
		}
	}
	return fallback
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, p := range split(s, sep) {
		p = trimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func split(s, sep string) []string {
	var result []string
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
