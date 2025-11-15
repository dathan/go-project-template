package config

import (
    "fmt"
    "os"
)

// Config holds application configuration loaded from environment variables.
// Each field corresponds to a service or dependency. This struct is kept
// minimal; additional configuration can be added as features evolve.
type Config struct {
    // HTTPPort is the port the API server listens on, e.g. ":8080".
    HTTPPort string
    // DatabaseURL is the PostgreSQL DSN, e.g. "postgres://user:pass@host:port/dbname".
    DatabaseURL string
    // RedisAddr is the Redis server address, e.g. "localhost:6379".
    RedisAddr string
    // RedisPassword is the Redis password if required.
    RedisPassword string
    // OAuth credentials for Google
    GoogleClientID     string
    GoogleClientSecret string
    // OAuth credentials for Slack
    SlackClientID     string
    SlackClientSecret string
    // OAuth credentials for LinkedIn
    LinkedInClientID     string
    LinkedInClientSecret string
    // OAuthRedirectURI is the redirect URI configured at the provider.
    OAuthRedirectURI string
    // SessionExpiration in seconds; sessions older than this will expire in Redis.
    SessionExpiration int
}

// Load reads configuration from environment variables. When a variable is not
// present the function falls back to a sensible default for local development.
func Load() (*Config, error) {
    cfg := &Config{
        HTTPPort:           getEnv("HTTP_PORT", ":8080"),
        DatabaseURL:        getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable"),
        RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword:      getEnv("REDIS_PASSWORD", ""),
        GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        SlackClientID:      os.Getenv("SLACK_CLIENT_ID"),
        SlackClientSecret:  os.Getenv("SLACK_CLIENT_SECRET"),
        LinkedInClientID:   os.Getenv("LINKEDIN_CLIENT_ID"),
        LinkedInClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
        OAuthRedirectURI:   getEnv("OAUTH_REDIRECT_URI", "http://localhost:8080/api/v1/auth/callback"),
        SessionExpiration:  getEnvInt("SESSION_EXPIRATION", 3600*24),
    }
    if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
        return nil, fmt.Errorf("missing Google OAuth credentials")
    }
    if cfg.SlackClientID == "" || cfg.SlackClientSecret == "" {
        return nil, fmt.Errorf("missing Slack OAuth credentials")
    }
    if cfg.LinkedInClientID == "" || cfg.LinkedInClientSecret == "" {
        return nil, fmt.Errorf("missing LinkedIn OAuth credentials")
    }
    return cfg, nil
}

func getEnv(key, defaultVal string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
    if v := os.Getenv(key); v != "" {
        var i int
        _, err := fmt.Sscanf(v, "%d", &i)
        if err == nil {
            return i
        }
    }
    return defaultVal
}