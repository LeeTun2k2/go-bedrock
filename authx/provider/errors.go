package provider

import "errors"

var (
	// ErrExchangeFailed indicates a failed provider token exchange.
	ErrExchangeFailed = errors.New("provider token exchange failed")
	// ErrInvalidProof indicates invalid provider identity proof.
	ErrInvalidProof = errors.New("invalid provider proof")
	// ErrInvalidRequest indicates an invalid auth request.
	ErrInvalidRequest = errors.New("invalid auth request")
)
