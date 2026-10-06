// Package github implements GitHub OAuth authentication.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/leetun2k2/go-bedrock/authx/provider"
)

const (
	defaultGitHubAuthURL        = "https://github.com/login/oauth/authorize"
	defaultGitHubTokenEndpoint  = "https://github.com/login/oauth/access_token"
	defaultGitHubUserEndpoint   = "https://api.github.com/user"
	defaultGitHubEmailsEndpoint = "https://api.github.com/user/emails"

	githubUserAgent = "authx"

	defaultGitHubHTTPTimeout = 10 * time.Second
)

// AuthorizationRequest is a type alias so consumers of this package do not
// need to import package provider directly.
type AuthorizationRequest = provider.AuthorizationRequest

// ExchangeRequest is a type alias so consumers of this package do not need
// to import package provider directly.
type ExchangeRequest = provider.ExchangeRequest

// Identity is a type alias so consumers of this package do not need to
// import package provider directly.
type Identity = provider.Identity

// Config contains configuration for the GitHub identity provider.
type Config struct {
	ClientID       string
	ClientSecret   string
	HTTPClient     *http.Client
	AuthURL        string // Optional override, defaults to GitHub OAuth auth endpoint
	TokenEndpoint  string // Optional override, defaults to GitHub token endpoint
	UserEndpoint   string // Optional override, defaults to GitHub user endpoint
	EmailsEndpoint string // Optional override, defaults to GitHub emails endpoint
}

// Provider implements provider.Provider for GitHub OAuth.
type Provider struct {
	clientID       string
	clientSecret   string
	httpClient     *http.Client
	authURL        string
	tokenEndpoint  string
	userEndpoint   string
	emailsEndpoint string
}

var _ provider.Provider = (*Provider)(nil)

// New creates a GitHub identity provider adapter.
func New(cfg Config) (*Provider, error) {
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, errors.New("github: client ID must not be empty")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("github: client secret must not be empty")
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: defaultGitHubHTTPTimeout}
	}

	authURL := cfg.AuthURL
	if authURL == "" {
		authURL = defaultGitHubAuthURL
	}

	tokenEndpoint := cfg.TokenEndpoint
	if tokenEndpoint == "" {
		tokenEndpoint = defaultGitHubTokenEndpoint
	}

	userEndpoint := cfg.UserEndpoint
	if userEndpoint == "" {
		userEndpoint = defaultGitHubUserEndpoint
	}

	emailsEndpoint := cfg.EmailsEndpoint
	if emailsEndpoint == "" {
		emailsEndpoint = defaultGitHubEmailsEndpoint
	}

	return &Provider{
		clientID:       cfg.ClientID,
		clientSecret:   cfg.ClientSecret,
		httpClient:     client,
		authURL:        authURL,
		tokenEndpoint:  tokenEndpoint,
		userEndpoint:   userEndpoint,
		emailsEndpoint: emailsEndpoint,
	}, nil
}

// AuthorizationURL constructs the GitHub authorization redirect URL.
func (g *Provider) AuthorizationURL(req provider.AuthorizationRequest) (string, error) {
	if strings.TrimSpace(req.RedirectURI) == "" || strings.TrimSpace(req.State) == "" {
		return "", provider.ErrInvalidRequest
	}

	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("redirect_uri", req.RedirectURI)
	q.Set("scope", "read:user user:email")
	q.Set("state", req.State)

	return g.authURL + "?" + q.Encode(), nil
}

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// Exchange exchanges the authorization code for an access token, fetches the
// user's profile and verified primary email, and returns a normalized Identity.
// Provider tokens are discarded.
func (g *Provider) Exchange(ctx context.Context, req provider.ExchangeRequest) (*provider.Identity, error) {
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.RedirectURI) == "" {
		return nil, provider.ErrInvalidRequest
	}

	accessToken, err := g.exchangeCode(ctx, req.Code, req.RedirectURI)
	if err != nil {
		return nil, err
	}

	user, err := g.fetchUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, provider.ErrInvalidProof
	}

	email, emailVerified, err := g.fetchPrimaryVerifiedEmail(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	emailMissing := strings.TrimSpace(email) == ""
	if emailMissing || !emailVerified {
		return nil, provider.ErrInvalidProof
	}

	displayName := user.Name
	if displayName == "" {
		displayName = user.Login
	}

	return &provider.Identity{
		Provider:      "github",
		Subject:       strconv.FormatInt(user.ID, 10),
		Email:         email,
		EmailVerified: emailVerified,
		DisplayName:   displayName,
		AvatarURL:     user.AvatarURL,
	}, nil
}

func (g *Provider) exchangeCode(ctx context.Context, code, redirectURI string) (string, error) {
	form := url.Values{}
	form.Set("client_id", g.clientID)
	form.Set("client_secret", g.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("%w: build request", provider.ErrExchangeFailed)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: call token endpoint", provider.ErrExchangeFailed)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%w: read response", provider.ErrExchangeFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return "", provider.ErrExchangeFailed
	}

	var tokenResp githubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("%w: decode response", provider.ErrExchangeFailed)
	}
	if tokenResp.Error != "" || tokenResp.AccessToken == "" {
		return "", provider.ErrExchangeFailed
	}

	return tokenResp.AccessToken, nil
}

func (g *Provider) fetchUser(ctx context.Context, accessToken string) (githubUser, error) {
	var user githubUser
	if err := g.doAuthenticatedGET(ctx, g.userEndpoint, accessToken, &user); err != nil {
		return githubUser{}, err
	}
	return user, nil
}

func (g *Provider) fetchPrimaryVerifiedEmail(ctx context.Context, accessToken string) (string, bool, error) {
	var emails []githubEmail
	if err := g.doAuthenticatedGET(ctx, g.emailsEndpoint, accessToken, &emails); err != nil {
		return "", false, err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, true, nil
		}
	}

	return "", false, nil
}

func (g *Provider) doAuthenticatedGET(ctx context.Context, endpoint, accessToken string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("%w: build request", provider.ErrExchangeFailed)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("User-Agent", githubUserAgent)
	httpReq.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: call api endpoint", provider.ErrExchangeFailed)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("%w: read response", provider.ErrExchangeFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return provider.ErrExchangeFailed
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%w: decode response", provider.ErrExchangeFailed)
	}

	return nil
}
