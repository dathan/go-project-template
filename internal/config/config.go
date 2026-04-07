// Package config loads application configuration from conf/config.yaml and .env.
// Priority (highest first): environment variables → .env file → conf/config.yaml → hard-coded defaults.
//
// Config file search order (first match wins):
//  1. <executable_dir>/../conf/config.yaml  (covers running from bin/)
//  2. <executable_dir>/conf/config.yaml
//  3. ./conf/config.yaml                    (covers go run ./cmd/server from project root)
//  4. ./ (last-resort)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config key constants — single source of truth used for SetDefault and GetString.
const (
	keyServerHost    = "server.host"
	keyServerPort    = "server.port"
	keyDBHost        = "database.host"
	keyDBPort        = "database.port"
	keyDBName        = "database.name"
	keyDBUser        = "database.user"
	keyDBPassword    = "database.password"
	keyDBSSLMode     = "database.sslmode"
	keyJWTSecret     = "auth.jwt_secret"
	keySessionDur    = "auth.session_duration"
	keyGoogleID      = "auth.oauth.google.client_id"
	keyGoogleSecret  = "auth.oauth.google.client_secret"
	keyGoogleRedir   = "auth.oauth.google.redirect_url"
	keyGitHubID      = "auth.oauth.github.client_id"
	keyGitHubSecret  = "auth.oauth.github.client_secret"
	keyGitHubRedir   = "auth.oauth.github.redirect_url"
	keySlackID       = "auth.oauth.slack.client_id"
	keySlackSecret   = "auth.oauth.slack.client_secret"
	keySlackRedir    = "auth.oauth.slack.redirect_url"
	keyLinkedInID    = "auth.oauth.linkedin.client_id"
	keyLinkedInSec   = "auth.oauth.linkedin.client_secret"
	keyLinkedInRedir = "auth.oauth.linkedin.redirect_url"
	keyStripeSecret  = "stripe.secret_key"
	keyStripeWebhook = "stripe.webhook_secret"
	keyStripePub     = "stripe.publishable_key"
	keyAgentProvider = "agent.provider"
	keyAgentModel    = "agent.model"
)

// Config is the root configuration struct.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Stripe   StripeConfig
	Agent    AgentConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// DSN returns a postgres connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode,
	)
}

type AuthConfig struct {
	JWTSecret       string
	SessionDuration time.Duration
	OAuth           OAuthConfig
}

type OAuthConfig struct {
	Google   OAuthProvider
	GitHub   OAuthProvider
	Slack    OAuthProvider
	LinkedIn OAuthProvider
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type StripeConfig struct {
	SecretKey      string
	WebhookSecret  string
	PublishableKey string
}

type AgentConfig struct {
	Provider string
	Model    string
}

// envOr returns the environment variable value when set and non-empty,
// otherwise returns fallback. This is the ONLY mechanism for env overrides —
// we do NOT use viper's AutomaticEnv or BindEnv because both have edge-case
// behaviours where an unset (or empty) env var can shadow a SetDefault value.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads conf/config.yaml (searched relative to the executable and cwd),
// then layers .env overrides, then explicit env var overrides.
func Load() (*Config, error) {
	// Load .env file relative to cwd — silently skip if absent.
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Search for conf/config.yaml relative to the running executable first.
	// This lets `./bin/server` find `<project>/conf/config.yaml` without
	// requiring the caller to cd to the project root.
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		v.AddConfigPath(filepath.Join(exeDir, "..", "conf")) // bin/../conf ✓
		v.AddConfigPath(filepath.Join(exeDir, "conf"))
	}
	v.AddConfigPath("./conf") // go run / test from project root
	v.AddConfigPath(".")

	// Hard-coded defaults — these win when neither config file nor env var
	// provides a value.
	v.SetDefault(keyServerHost, "0.0.0.0")
	v.SetDefault(keyServerPort, 8080)
	v.SetDefault(keyDBHost, "127.0.0.1")
	v.SetDefault(keyDBPort, 5432)
	v.SetDefault(keyDBName, "app_db")
	v.SetDefault(keyDBUser, "app_user")
	v.SetDefault(keyDBSSLMode, "disable")
	v.SetDefault(keySessionDur, "24h")
	v.SetDefault(keyAgentProvider, "claude")
	v.SetDefault(keyAgentModel, "claude-opus-4-6")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// Config file not found — defaults and env vars are sufficient.
	}

	sessionDuration, err := time.ParseDuration(v.GetString(keySessionDur))
	if err != nil {
		sessionDuration = 24 * time.Hour
	}

	// Build config: env var > config file > SetDefault.
	// envOr() handles the env var layer; v.GetString() covers file + default.
	cfg := &Config{
		Server: ServerConfig{
			Host: envOr("SERVER_HOST", v.GetString(keyServerHost)),
			Port: v.GetInt(keyServerPort),
		},
		Database: DatabaseConfig{
			Host:     envOr("DATABASE_HOST", v.GetString(keyDBHost)),
			Port:     v.GetInt(keyDBPort),
			Name:     envOr("DATABASE_NAME", v.GetString(keyDBName)),
			User:     envOr("DATABASE_USER", v.GetString(keyDBUser)),
			Password: envOr("DATABASE_PASSWORD", v.GetString(keyDBPassword)),
			SSLMode:  envOr("DATABASE_SSLMODE", v.GetString(keyDBSSLMode)),
		},
		Auth: AuthConfig{
			JWTSecret:       envOr("AUTH_JWT_SECRET", v.GetString(keyJWTSecret)),
			SessionDuration: sessionDuration,
			OAuth: OAuthConfig{
				Google: OAuthProvider{
					ClientID:     envOr("AUTH_OAUTH_GOOGLE_CLIENT_ID", v.GetString(keyGoogleID)),
					ClientSecret: envOr("AUTH_OAUTH_GOOGLE_CLIENT_SECRET", v.GetString(keyGoogleSecret)),
					RedirectURL:  envOr("AUTH_OAUTH_GOOGLE_REDIRECT_URL", v.GetString(keyGoogleRedir)),
				},
				GitHub: OAuthProvider{
					ClientID:     envOr("AUTH_OAUTH_GITHUB_CLIENT_ID", v.GetString(keyGitHubID)),
					ClientSecret: envOr("AUTH_OAUTH_GITHUB_CLIENT_SECRET", v.GetString(keyGitHubSecret)),
					RedirectURL:  envOr("AUTH_OAUTH_GITHUB_REDIRECT_URL", v.GetString(keyGitHubRedir)),
				},
				Slack: OAuthProvider{
					ClientID:     envOr("AUTH_OAUTH_SLACK_CLIENT_ID", v.GetString(keySlackID)),
					ClientSecret: envOr("AUTH_OAUTH_SLACK_CLIENT_SECRET", v.GetString(keySlackSecret)),
					RedirectURL:  envOr("AUTH_OAUTH_SLACK_REDIRECT_URL", v.GetString(keySlackRedir)),
				},
				LinkedIn: OAuthProvider{
					ClientID:     envOr("AUTH_OAUTH_LINKEDIN_CLIENT_ID", v.GetString(keyLinkedInID)),
					ClientSecret: envOr("AUTH_OAUTH_LINKEDIN_CLIENT_SECRET", v.GetString(keyLinkedInSec)),
					RedirectURL:  envOr("AUTH_OAUTH_LINKEDIN_REDIRECT_URL", v.GetString(keyLinkedInRedir)),
				},
			},
		},
		Stripe: StripeConfig{
			SecretKey:      envOr("STRIPE_SECRET_KEY", v.GetString(keyStripeSecret)),
			WebhookSecret:  envOr("STRIPE_WEBHOOK_SECRET", v.GetString(keyStripeWebhook)),
			PublishableKey: envOr("STRIPE_PUBLISHABLE_KEY", v.GetString(keyStripePub)),
		},
		Agent: AgentConfig{
			Provider: envOr("AGENT_PROVIDER", v.GetString(keyAgentProvider)),
			Model:    envOr("AGENT_MODEL", v.GetString(keyAgentModel)),
		},
	}

	return cfg, nil
}
