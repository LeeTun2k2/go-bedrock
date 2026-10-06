// Package oauth provides OAuth state, nonce, code, and PKCE helpers.
package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// PKCEMethodS256 is the only supported PKCE method; plain PKCE and others are rejected.
const PKCEMethodS256 = "S256"

// defaultTokenBytes is the default byte length for generated state, nonce, and code tokens.
// 32 bytes (256 bits) exceeds the 128-bit minimum required by OAuth 2.0 / OIDC specs.
const defaultTokenBytes = 32

// ComputeCodeChallenge derives the S256 PKCE code challenge from a verifier.
// Returns unpadded base64url encoding of the SHA-256 digest.
func ComputeCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// VerifyCodeChallenge verifies whether verifier matches challenge under method.
// Any method other than S256 fails closed.
func VerifyCodeChallenge(verifier, challenge, method string) bool {
	if method != PKCEMethodS256 {
		return false
	}
	computed := ComputeCodeChallenge(verifier)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

// GenerateRandomToken generates a cryptographically random, base64url-encoded token
// of the specified byte length.
func GenerateRandomToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = defaultTokenBytes
	}
	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// GenerateState generates a random OAuth state string for CSRF mitigation.
func GenerateState() (string, error) {
	return GenerateRandomToken(defaultTokenBytes)
}

// GenerateNonce generates a random OIDC nonce string for replay mitigation.
func GenerateNonce() (string, error) {
	return GenerateRandomToken(defaultTokenBytes)
}

// GenerateAuthorizationCode generates a random one-time authorization code.
func GenerateAuthorizationCode() (string, error) {
	return GenerateRandomToken(defaultTokenBytes)
}
