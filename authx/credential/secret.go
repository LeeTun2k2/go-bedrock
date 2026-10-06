package credential

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters following OWASP baseline recommendations for interactive logins.
const (
	argon2Time    uint32 = 1
	argon2Memory  uint32 = 64 * 1024 // 64 MiB
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
	saltLen              = 16
)

// HashSecret hashes a plaintext client secret with Argon2id and returns a
// self-describing PHC-style encoded string. Plaintext is never stored or returned.
func HashSecret(secret string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(secret), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Time,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifySecret reports whether the plaintext secret matches a hash produced by HashSecret.
func VerifySecret(secret, encoded string) (bool, error) {
	var (
		version              int
		memory, time, thread uint32
	)

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid argon2id hash format")
	}

	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parse argon2id version: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("unsupported argon2id version %d", version)
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &thread); err != nil {
		return false, fmt.Errorf("parse argon2id parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode argon2id salt: %w", err)
	}

	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode argon2id hash: %w", err)
	}

	got := argon2.IDKey([]byte(secret), salt, time, memory, uint8(thread), uint32(len(want)))

	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// VerifyClientSecret verifies a client secret against an Argon2id hash, collapsing
// all validation and format errors into ErrInvalidClient to prevent leaking details.
func VerifyClientSecret(secret, hash string) error {
	ok, err := VerifySecret(secret, hash)
	if err != nil || !ok {
		return ErrInvalidClient
	}
	return nil
}
