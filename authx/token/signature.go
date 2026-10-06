package token

import (
	"encoding/base64"
	"errors"
	"math/big"
)

func decodeBase64URLSegment(segment string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(segment)
}

func encodeBase64URLSegment(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func encodeECDSASignature(r, s *big.Int) ([]byte, error) {
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	componentTooLarge := len(rBytes) > ecdsaP256FieldBytes || len(sBytes) > ecdsaP256FieldBytes
	if componentTooLarge {
		return nil, errors.New("ECDSA signature component too large for P-256")
	}

	signature := make([]byte, 2*ecdsaP256FieldBytes)
	copy(signature[ecdsaP256FieldBytes-len(rBytes):ecdsaP256FieldBytes], rBytes)
	copy(signature[2*ecdsaP256FieldBytes-len(sBytes):], sBytes)
	return signature, nil
}

func decodeECDSASignature(signature []byte) (*big.Int, *big.Int, error) {
	if len(signature) != 2*ecdsaP256FieldBytes {
		return nil, nil, errors.New("invalid ECDSA signature length")
	}

	r := new(big.Int).SetBytes(signature[:ecdsaP256FieldBytes])
	s := new(big.Int).SetBytes(signature[ecdsaP256FieldBytes:])
	return r, s, nil
}
