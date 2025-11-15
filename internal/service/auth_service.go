package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/google/uuid"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"

    "github.com/dathan/go-project-template/internal/config"
    "github.com/dathan/go-project-template/internal/model"
    pgstore "github.com/dathan/go-project-template/internal/store/postgres"
    redistore "github.com/dathan/go-project-template/internal/store/redis"
)

// AuthService orchestrates the OAuth2 login flow across multiple providers.
// It relies on per-provider definitions implementing the provider interface
// below to abstract differences in user info APIs.
type AuthService struct {
    cfg          *config.Config
    userRepo     pgstore.UserRepository
    sessionStore redistore.SessionStore
    providers    map[string]OAuthProvider
    sessionTTL   time.Duration
}

// OAuthProvider abstracts a third-party identity provider. Each provider
// returns an OAuth2 configuration and knows how to retrieve basic user
// information using an access token.
type OAuthProvider interface {
    Config() *oauth2.Config
    GetUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (id, email, name string, err error)
}

// NewAuthService constructs a new AuthService. It registers built-in
// providers (Google, Slack, LinkedIn) based on supplied configuration.
func NewAuthService(cfg *config.Config, userRepo pgstore.UserRepository, sessionStore redistore.SessionStore) (*AuthService, error) {
    providers := map[string]OAuthProvider{}
    // Google provider
    providers["google"] = &googleProvider{
        clientID:     cfg.GoogleClientID,
        clientSecret: cfg.GoogleClientSecret,
        redirectURI:  cfg.OAuthRedirectURI,
    }
    // Slack provider
    providers["slack"] = &slackProvider{
        clientID:     cfg.SlackClientID,
        clientSecret: cfg.SlackClientSecret,
        redirectURI:  cfg.OAuthRedirectURI,
    }
    // LinkedIn provider
    providers["linkedin"] = &linkedInProvider{
        clientID:     cfg.LinkedInClientID,
        clientSecret: cfg.LinkedInClientSecret,
        redirectURI:  cfg.OAuthRedirectURI,
    }
    return &AuthService{
        cfg:          cfg,
        userRepo:     userRepo,
        sessionStore: sessionStore,
        providers:    providers,
        sessionTTL:   time.Duration(cfg.SessionExpiration) * time.Second,
    }, nil
}

// SessionTTL returns the configured session expiration duration. This
// method is public so HTTP handlers can construct cookies with an
// appropriate expiry.
func (s *AuthService) SessionTTL() time.Duration {
    return s.sessionTTL
}

// GetLoginURL returns the OAuth2 authorization URL for a provider. The
// state parameter is not persisted since CSRF protection requires a more
// sophisticated implementation (outside scope). This method may return an
// error if the provider is unknown.
func (s *AuthService) GetLoginURL(provider string) (string, error) {
    p, ok := s.providers[provider]
    if !ok {
        return "", fmt.Errorf("unsupported provider: %s", provider)
    }
    authURL := p.Config().AuthCodeURL("", oauth2.AccessTypeOffline)
    return authURL, nil
}

// HandleCallback processes the OAuth callback. It exchanges the code for
// tokens, retrieves user info, persists the user and creates a session.
// It returns a session ID which should be set in an HTTP-only cookie.
func (s *AuthService) HandleCallback(ctx context.Context, provider, code string) (string, error) {
    p, ok := s.providers[provider]
    if !ok {
        return "", fmt.Errorf("unsupported provider: %s", provider)
    }
    conf := p.Config()
    token, err := conf.Exchange(ctx, code)
    if err != nil {
        return "", fmt.Errorf("token exchange: %w", err)
    }
    client := conf.Client(ctx, token)
    id, email, name, err := p.GetUserInfo(ctx, client, token)
    if err != nil {
        return "", fmt.Errorf("get user info: %w", err)
    }
    // Create or update user in the database
    user := &model.User{
        Provider:   provider,
        ProviderID: id,
        Email:      email,
        Name:       name,
    }
    if _, err := s.userRepo.CreateOrUpdate(ctx, user); err != nil {
        return "", fmt.Errorf("persist user: %w", err)
    }
    // Generate session ID and store session data in Redis
    sessionID := uuid.NewString()
    tokenData := map[string]string{
        "access_token":  token.AccessToken,
    }
    if token.RefreshToken != "" {
        tokenData["refresh_token"] = token.RefreshToken
    }
    data := redistore.SessionData{
        UserID:    user.ID,
        Provider:  provider,
        TokenData: tokenData,
    }
    if err := s.sessionStore.Set(ctx, sessionID, data, s.sessionTTL); err != nil {
        return "", fmt.Errorf("store session: %w", err)
    }
    return sessionID, nil
}

// --- Provider Implementations ---

// googleProvider implements OAuthProvider for Google.
type googleProvider struct {
    clientID     string
    clientSecret string
    redirectURI  string
}

func (p *googleProvider) Config() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     p.clientID,
        ClientSecret: p.clientSecret,
        RedirectURL:  p.redirectURI,
        Scopes:       []string{"openid", "profile", "email"},
        Endpoint:     google.Endpoint,
    }
}

func (p *googleProvider) GetUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (string, string, string, error) {
    // Google user info endpoint
    const userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
    resp, err := client.Get(userInfoURL)
    if err != nil {
        return "", "", "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", "", "", fmt.Errorf("google userinfo returned status %s", resp.Status)
    }
    var data struct {
        Sub   string `json:"sub"`
        Email string `json:"email"`
        Name  string `json:"name"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return "", "", "", err
    }
    return data.Sub, data.Email, data.Name, nil
}

// slackProvider implements OAuthProvider for Slack.
type slackProvider struct {
    clientID     string
    clientSecret string
    redirectURI  string
}

func (p *slackProvider) Config() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     p.clientID,
        ClientSecret: p.clientSecret,
        RedirectURL:  p.redirectURI,
        Scopes:       []string{"identity.basic", "identity.email", "identity.avatar", "identity.team"},
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://slack.com/oauth/v2/authorize",
            TokenURL: "https://slack.com/api/oauth.v2.access",
        },
    }
}

func (p *slackProvider) GetUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (string, string, string, error) {
    // Slack user identity endpoint
    const userInfoURL = "https://slack.com/api/users.identity"
    resp, err := client.Get(userInfoURL)
    if err != nil {
        return "", "", "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", "", "", fmt.Errorf("slack identity returned status %s", resp.Status)
    }
    var data struct {
        OK    bool `json:"ok"`
        User  struct {
            ID    string `json:"id"`
            Name  string `json:"name"`
            Email string `json:"email"`
        } `json:"user"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return "", "", "", err
    }
    if !data.OK {
        return "", "", "", errors.New("slack identity returned not ok")
    }
    return data.User.ID, data.User.Email, data.User.Name, nil
}

// linkedInProvider implements OAuthProvider for LinkedIn.
type linkedInProvider struct {
    clientID     string
    clientSecret string
    redirectURI  string
}

func (p *linkedInProvider) Config() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     p.clientID,
        ClientSecret: p.clientSecret,
        RedirectURL:  p.redirectURI,
        Scopes:       []string{"r_liteprofile", "r_emailaddress"},
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
            TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
        },
    }
}

func (p *linkedInProvider) GetUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (string, string, string, error) {
    // LinkedIn user info requires two API calls: basic profile and email address
    // Basic profile
    profResp, err := client.Get("https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName)")
    if err != nil {
        return "", "", "", err
    }
    defer profResp.Body.Close()
    if profResp.StatusCode != http.StatusOK {
        return "", "", "", fmt.Errorf("linkedin profile status %s", profResp.Status)
    }
    var profile struct {
        ID        string `json:"id"`
        FirstName struct {
            Localized map[string]string `json:"localized"`
        } `json:"localizedFirstName"`
        LastName struct {
            Localized map[string]string `json:"localized"`
        } `json:"localizedLastName"`
    }
    if err := json.NewDecoder(profResp.Body).Decode(&profile); err != nil {
        return "", "", "", err
    }
    // Email address
    emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
    if err != nil {
        return "", "", "", err
    }
    defer emailResp.Body.Close()
    if emailResp.StatusCode != http.StatusOK {
        return "", "", "", fmt.Errorf("linkedin email status %s", emailResp.Status)
    }
    var emailData struct {
        Elements []struct {
            Handle string `json:"handle"`
            HandleTilde struct {
                EmailAddress string `json:"emailAddress"`
            } `json:"handle~"`
        } `json:"elements"`
    }
    if err := json.NewDecoder(emailResp.Body).Decode(&emailData); err != nil {
        return "", "", "", err
    }
    email := ""
    if len(emailData.Elements) > 0 {
        email = emailData.Elements[0].HandleTilde.EmailAddress
    }
    // Compose full name
    first := ""
    if v, ok := profile.FirstName.Localized["en_US"]; ok {
        first = v
    }
    last := ""
    if v, ok := profile.LastName.Localized["en_US"]; ok {
        last = v
    }
    fullName := first + " " + last
    return profile.ID, email, fullName, nil
}