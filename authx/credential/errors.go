// Package credential handles client secrets and HTTP client credentials.
package credential

import "errors"

// ErrInvalidClient indicates invalid client credentials.
var ErrInvalidClient = errors.New("invalid client credentials")
