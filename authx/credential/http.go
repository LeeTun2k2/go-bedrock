package credential

import (
	"net/http"
	"strings"
)

const bearerPrefix = "bearer "

// Credentials holds client_id and client_secret extracted from an HTTP request.
type Credentials struct {
	ClientID     string
	ClientSecret string
}

// Extract extracts client credentials from either the HTTP Basic
// Auth header or the X-Client-Id / X-Client-Secret header pair.
// Supplying both forms, an incomplete form, or no credentials fails closed with ErrInvalidClient.
func Extract(r *http.Request) (Credentials, error) {
	if r == nil {
		return Credentials{}, ErrInvalidClient
	}

	authHeader := r.Header.Get("Authorization")
	hasBasic := strings.HasPrefix(strings.ToLower(authHeader), "basic ")

	headerID := r.Header.Get("X-Client-Id")
	headerSecret := r.Header.Get("X-Client-Secret")
	hasHeader := headerID != "" || headerSecret != ""

	if hasBasic && hasHeader {
		return Credentials{}, ErrInvalidClient
	}

	switch {
	case hasBasic:
		id, secret, ok := r.BasicAuth()
		if !ok || strings.TrimSpace(id) == "" || strings.TrimSpace(secret) == "" {
			return Credentials{}, ErrInvalidClient
		}
		return Credentials{ClientID: id, ClientSecret: secret}, nil
	case hasHeader:
		if strings.TrimSpace(headerID) == "" || strings.TrimSpace(headerSecret) == "" {
			return Credentials{}, ErrInvalidClient
		}
		return Credentials{ClientID: headerID, ClientSecret: headerSecret}, nil
	default:
		return Credentials{}, ErrInvalidClient
	}
}

// ExtractBearerToken extracts a bearer token from an Authorization header value.
// The scheme check is case-insensitive per RFC 7235. Returns false if the header
// is missing, does not start with Bearer, or has an empty token.
func ExtractBearerToken(header string) (string, bool) {
	if len(header) < len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(bearerPrefix):])
	if token == "" {
		return "", false
	}

	return token, true
}

// ExtractBearerTokenFromRequest extracts a bearer token from the Authorization
// header of an HTTP request.
func ExtractBearerTokenFromRequest(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	return ExtractBearerToken(r.Header.Get("Authorization"))
}
