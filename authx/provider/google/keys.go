package google

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/leetun2k2/go-bedrock/authx/provider"
)

type googleJWKS struct {
	Keys []googleJWK `json:"keys"`
}

type googleJWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (g *Provider) lookupKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	g.keysMu.Lock()
	defer g.keysMu.Unlock()

	cacheIsFresh := time.Since(g.keysFetched) < defaultGoogleKeysCacheTTL
	key, keyExists := g.keys[kid]
	if keyExists && cacheIsFresh {
		return key, nil
	}

	keys, err := fetchGoogleKeys(ctx, g.httpClient, g.certsEndpoint)
	if err != nil {
		return nil, err
	}

	g.keys = keys
	g.keysFetched = time.Now()

	key, keyExists = g.keys[kid]
	if !keyExists {
		return nil, provider.ErrInvalidProof
	}

	return key, nil
}

func fetchGoogleKeys(
	ctx context.Context,
	httpClient *http.Client,
	certsEndpoint string,
) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build certs request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call certs endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read certs response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("certs endpoint returned status %d", resp.StatusCode)
	}

	var jwks googleJWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("decode certs response: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, jwk := range jwks.Keys {
		isRSAKey := jwk.Kty == "RSA"
		if !isRSAKey || jwk.Kid == "" {
			continue
		}

		publicKey, err := rsaPublicKeyFromJWK(jwk.N, jwk.E)
		if err != nil {
			continue
		}
		keys[jwk.Kid] = publicKey
	}

	return keys, nil
}

func rsaPublicKeyFromJWK(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := decodeBase64URLSegment(modulus)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}

	eBytes, err := decodeBase64URLSegment(exponent)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}

func decodeBase64URLSegment(segment string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(segment)
}

func encodeBase64URLSegment(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
