package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/dathan/go-project-template/internal/config"
)

// UserInfo is the normalised profile returned by every provider.
type UserInfo struct {
	Provider   string
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

// OAuthProvider wraps an oauth2.Config and knows how to fetch the user profile.
type OAuthProvider struct {
	cfg      *oauth2.Config
	fetchFn  func(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}

// AuthCodeURL returns the redirect URL and a CSRF state token.
func (p *OAuthProvider) AuthCodeURL() (url, state string, err error) {
	b := make([]byte, 16)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating state: %w", err)
	}
	state = hex.EncodeToString(b)
	url = p.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return url, state, nil
}

// Exchange trades the authorization code for a user profile.
func (p *OAuthProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	token, err := p.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	return p.fetchFn(ctx, token)
}

// Registry holds all configured OAuth providers.
type Registry struct {
	providers map[string]*OAuthProvider
}

// NewRegistry builds a provider registry from config.
func NewRegistry(cfg config.OAuthConfig) *Registry {
	r := &Registry{providers: make(map[string]*OAuthProvider)}

	if cfg.Google.ClientID != "" {
		r.providers["google"] = &OAuthProvider{
			cfg: &oauth2.Config{
				ClientID:     cfg.Google.ClientID,
				ClientSecret: cfg.Google.ClientSecret,
				RedirectURL:  cfg.Google.RedirectURL,
				Scopes:       []string{"openid", "email", "profile"},
				Endpoint:     google.Endpoint,
			},
			fetchFn: fetchGoogle,
		}
	}

	if cfg.GitHub.ClientID != "" {
		r.providers["github"] = &OAuthProvider{
			cfg: &oauth2.Config{
				ClientID:     cfg.GitHub.ClientID,
				ClientSecret: cfg.GitHub.ClientSecret,
				RedirectURL:  cfg.GitHub.RedirectURL,
				Scopes:       []string{"user:email", "read:user"},
				Endpoint:     github.Endpoint,
			},
			fetchFn: fetchGitHub,
		}
	}

	if cfg.Slack.ClientID != "" {
		r.providers["slack"] = &OAuthProvider{
			cfg: &oauth2.Config{
				ClientID:     cfg.Slack.ClientID,
				ClientSecret: cfg.Slack.ClientSecret,
				RedirectURL:  cfg.Slack.RedirectURL,
				Scopes:       []string{"users:read", "users:read.email"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://slack.com/oauth/v2/authorize",
					TokenURL: "https://slack.com/api/oauth.v2.access",
				},
			},
			fetchFn: fetchSlack,
		}
	}

	if cfg.LinkedIn.ClientID != "" {
		r.providers["linkedin"] = &OAuthProvider{
			cfg: &oauth2.Config{
				ClientID:     cfg.LinkedIn.ClientID,
				ClientSecret: cfg.LinkedIn.ClientSecret,
				RedirectURL:  cfg.LinkedIn.RedirectURL,
				Scopes:       []string{"openid", "profile", "email"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
					TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
				},
			},
			fetchFn: fetchLinkedIn,
		}
	}

	return r
}

// Get returns the provider for the given name, or an error if unknown/unconfigured.
func (r *Registry) Get(name string) (*OAuthProvider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown or unconfigured oauth provider: %s", name)
	}
	return p, nil
}

// ── Provider-specific profile fetchers ────────────────────────────────────────

func fetchGoogle(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	data, err := getJSON(ctx, "https://www.googleapis.com/oauth2/v3/userinfo", token)
	if err != nil {
		return nil, err
	}
	return &UserInfo{
		Provider:   "google",
		ProviderID: str(data["sub"]),
		Email:      str(data["email"]),
		Name:       str(data["name"]),
		AvatarURL:  str(data["picture"]),
	}, nil
}

func fetchGitHub(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	data, err := getJSON(ctx, "https://api.github.com/user", token)
	if err != nil {
		return nil, err
	}
	// GitHub may not expose email in /user; fetch from /user/emails if needed.
	email := str(data["email"])
	if email == "" {
		email, _ = fetchGitHubEmail(ctx, token)
	}
	return &UserInfo{
		Provider:   "github",
		ProviderID: fmt.Sprintf("%v", data["id"]),
		Email:      email,
		Name:       str(data["name"]),
		AvatarURL:  str(data["avatar_url"]),
	}, nil
}

func fetchGitHubEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var emails []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if primary, _ := e["primary"].(bool); primary {
			return str(e["email"]), nil
		}
	}
	return "", nil
}

func fetchSlack(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	data, err := getJSON(ctx, "https://slack.com/api/users.identity", token)
	if err != nil {
		return nil, err
	}
	user, _ := data["user"].(map[string]any)
	return &UserInfo{
		Provider:   "slack",
		ProviderID: str(user["id"]),
		Email:      str(user["email"]),
		Name:       str(user["name"]),
		AvatarURL:  str(user["image_72"]),
	}, nil
}

func fetchLinkedIn(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	data, err := getJSON(ctx, "https://api.linkedin.com/v2/userinfo", token)
	if err != nil {
		return nil, err
	}
	return &UserInfo{
		Provider:   "linkedin",
		ProviderID: str(data["sub"]),
		Email:      str(data["email"]),
		Name:       str(data["name"]),
		AvatarURL:  str(data["picture"]),
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func getJSON(ctx context.Context, url string, token *oauth2.Token) (map[string]any, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provider returned %d: %s", resp.StatusCode, body)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return out, nil
}

func str(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
