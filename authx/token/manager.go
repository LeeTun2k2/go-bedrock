package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	joseAlgES256 = "ES256"
	joseTypJWT   = "JWT"

	// ecdsaP256FieldBytes is the fixed byte width of each P-256 scalar in a
	// JWS ES256 signature (32 bytes for r, 32 bytes for s).
	ecdsaP256FieldBytes = 32

	// clockSkew is the allowance for iat validation against clock drift.
	clockSkew = 1 * time.Minute
)

// Key holds a key identifier and ECDSA P-256 keys.
type Key struct {
	ID         string
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// Config holds the configuration for Manager.
type Config struct {
	Issuer           string
	ActiveKey        Key
	VerificationKeys []Key
}

// Manager signs and verifies ES256 session JWTs with kid, issuer, audience,
// expiry checks, and rotation overlap support.
type Manager struct {
	issuer           string
	activeKeyID      string
	activePrivateKey *ecdsa.PrivateKey
	verificationKeys map[string]*ecdsa.PublicKey
}

// NewManager creates a Manager configured with an active signing key
// and any previous verification keys for key rotation overlap.
func NewManager(cfg Config) (*Manager, error) {
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, errors.New("authx: token issuer must not be empty")
	}
	if strings.TrimSpace(cfg.ActiveKey.ID) == "" {
		return nil, errors.New("authx: active key ID must not be empty")
	}
	if cfg.ActiveKey.PrivateKey == nil {
		return nil, errors.New("authx: active key private key must not be nil")
	}
	if cfg.ActiveKey.PrivateKey.Curve != elliptic.P256() {
		return nil, errors.New("authx: active key must use curve P-256")
	}

	keys := make(map[string]*ecdsa.PublicKey, len(cfg.VerificationKeys)+1)

	for _, vk := range cfg.VerificationKeys {
		if strings.TrimSpace(vk.ID) == "" {
			return nil, errors.New("authx: verification key ID must not be empty")
		}
		pub := vk.PublicKey
		if pub == nil && vk.PrivateKey != nil {
			pub = &vk.PrivateKey.PublicKey
		}
		if pub == nil {
			return nil, fmt.Errorf("authx: verification key %q must have a public key", vk.ID)
		}
		if pub.Curve != elliptic.P256() {
			return nil, fmt.Errorf("authx: verification key %q must use curve P-256", vk.ID)
		}
		keys[vk.ID] = pub
	}

	activePub := cfg.ActiveKey.PublicKey
	if activePub == nil {
		activePub = &cfg.ActiveKey.PrivateKey.PublicKey
	}
	keys[cfg.ActiveKey.ID] = activePub

	return &Manager{
		issuer:           cfg.Issuer,
		activeKeyID:      cfg.ActiveKey.ID,
		activePrivateKey: cfg.ActiveKey.PrivateKey,
		verificationKeys: keys,
	}, nil
}

// Issuer returns the configured token issuer.
func (tm *Manager) Issuer() string {
	return tm.issuer
}

type joseHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

// Sign produces a compact ES256 JWT for claims using the active signing key.
func (tm *Manager) Sign(claims Claims) (string, error) {
	header := joseHeader{
		Alg: joseAlgES256,
		Typ: joseTypJWT,
		Kid: tm.activeKeyID,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal jose header: %w", err)
	}

	if claims.Issuer == "" {
		claims.Issuer = tm.issuer
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	signingInput := encodeBase64URLSegment(headerJSON) + "." + encodeBase64URLSegment(payloadJSON)

	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, tm.activePrivateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	signature, err := encodeECDSASignature(r, s)
	if err != nil {
		return "", fmt.Errorf("encode signature: %w", err)
	}

	return signingInput + "." + encodeBase64URLSegment(signature), nil
}

// Verify parses and validates a compact JWT against the configured verification
// keys, issuer, and required audience.
func (tm *Manager) Verify(token, expectedAudience string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	headerPart, payloadPart, signaturePart := parts[0], parts[1], parts[2]

	headerJSON, err := decodeBase64URLSegment(headerPart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header joseHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}
	if header.Alg != joseAlgES256 {
		return nil, ErrUnsupportedAlgorithm
	}

	publicKey, ok := tm.verificationKeys[header.Kid]
	if !ok {
		return nil, ErrUnknownKey
	}

	signature, err := decodeBase64URLSegment(signaturePart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	r, s, err := decodeECDSASignature(signature)
	if err != nil {
		return nil, ErrInvalidToken
	}

	hash := sha256.Sum256([]byte(headerPart + "." + payloadPart))
	if !ecdsa.Verify(publicKey, hash[:], r, s) {
		return nil, ErrInvalidToken
	}

	payloadJSON, err := decodeBase64URLSegment(payloadPart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if err := tm.validateClaims(claims, expectedAudience); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (tm *Manager) validateClaims(claims Claims, expectedAudience string) error {
	if strings.TrimSpace(expectedAudience) == "" {
		return ErrInvalidToken
	}
	if claims.Issuer != tm.issuer || claims.Audience == "" {
		return ErrInvalidToken
	}
	if claims.Audience != expectedAudience {
		return ErrInvalidToken
	}

	now := time.Now()
	if !time.Unix(claims.ExpiresAt, 0).After(now) {
		return ErrInvalidToken
	}
	if time.Unix(claims.IssuedAt, 0).After(now.Add(clockSkew)) {
		return ErrInvalidToken
	}

	return nil
}
