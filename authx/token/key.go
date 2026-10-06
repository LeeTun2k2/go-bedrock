package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// GenerateSigningKey creates a new in-memory P-256 private key, intended
// for testing and local development.
func GenerateSigningKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ES256 signing key: %w", err)
	}
	return key, nil
}

// ParseECPrivateKeyPEM parses a PEM-encoded EC private key in either
// SEC1 ("EC PRIVATE KEY") or PKCS8 ("PRIVATE KEY") form.
func ParseECPrivateKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("decode PEM block: no PEM data found")
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC private key: %w", err)
	}

	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("parse EC private key: not an ECDSA key")
	}
	return key, nil
}

// ParseECPublicKeyPEM parses a PEM-encoded PKIX EC public key.
func ParseECPublicKeyPEM(pemBytes []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("decode PEM block: no PEM data found")
	}

	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC public key: %w", err)
	}

	key, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("parse EC public key: not an ECDSA key")
	}
	return key, nil
}
