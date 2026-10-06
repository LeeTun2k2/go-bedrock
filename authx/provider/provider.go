// Package provider defines shared identity-provider contracts.
package provider

import (
	"context"
	"strings"
)

// NormalizeEmail trims surrounding whitespace and lowercases email for stable
// canonical comparison and uniqueness checks. Returns "" for empty or whitespace-only inputs.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Identity is the normalized external identity returned by identity providers.
// Provider tokens and raw payloads are discarded and never cross package boundaries.
type Identity struct {
	Provider      string `json:"provider"`
	Subject       string `json:"subject"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	DisplayName   string `json:"display_name,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
}

// AuthorizationRequest contains the parameters needed to construct a provider
// authorization redirect URL.
type AuthorizationRequest struct {
	RedirectURI string
	State       string
	Nonce       string // Required for OIDC providers like Google; optional/ignored for GitHub
}

// ExchangeRequest contains the parameters required to exchange a provider
// authorization code for trusted identity claims.
type ExchangeRequest struct {
	Code        string
	RedirectURI string
	Nonce       string // Required for OIDC providers like Google; optional/ignored for GitHub
}

// Provider represents an external identity provider capable of building
// authorization redirect URLs and exchanging authorization codes for normalized identities.
type Provider interface {
	AuthorizationURL(req AuthorizationRequest) (string, error)
	Exchange(ctx context.Context, req ExchangeRequest) (*Identity, error)
}
