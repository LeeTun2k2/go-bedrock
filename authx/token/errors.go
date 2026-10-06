// Package token signs and verifies ES256 session tokens.
package token

import "errors"

var (
	// ErrInvalidToken indicates an invalid session token.
	ErrInvalidToken = errors.New("invalid token")
	// ErrUnknownKey indicates an unknown signing key.
	ErrUnknownKey = errors.New("unknown signing key")
	// ErrUnsupportedAlgorithm indicates an unsupported token algorithm.
	ErrUnsupportedAlgorithm = errors.New("unsupported token algorithm")
)
