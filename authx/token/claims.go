package token

import "encoding/json"

// Claims represents registered JWT claims with service-defined claims.
type Claims struct {
	Issuer    string         `json:"iss"`
	Audience  string         `json:"aud"`
	Subject   string         `json:"sub"`
	IssuedAt  int64          `json:"iat"`
	ExpiresAt int64          `json:"exp"`
	Custom    map[string]any `json:"-"`
}

// MarshalJSON combines registered and custom claims.
func (c Claims) MarshalJSON() ([]byte, error) {
	claims := make(map[string]any, len(c.Custom)+5)
	for key, value := range c.Custom {
		claims[key] = value
	}

	setRegisteredClaims(claims, c)
	return json.Marshal(claims)
}

func setRegisteredClaims(claims map[string]any, registered Claims) {
	if registered.Issuer != "" {
		claims["iss"] = registered.Issuer
	}
	if registered.Audience != "" {
		claims["aud"] = registered.Audience
	}
	if registered.Subject != "" {
		claims["sub"] = registered.Subject
	}
	if registered.IssuedAt != 0 {
		claims["iat"] = registered.IssuedAt
	}
	if registered.ExpiresAt != 0 {
		claims["exp"] = registered.ExpiresAt
	}
}

// UnmarshalJSON separates registered claims from custom claims.
func (c *Claims) UnmarshalJSON(data []byte) error {
	var claims map[string]any
	if err := json.Unmarshal(data, &claims); err != nil {
		return err
	}

	c.readRegisteredClaims(claims)
	c.Custom = claims
	return nil
}

func (c *Claims) readRegisteredClaims(claims map[string]any) {
	if issuer, ok := claims["iss"].(string); ok {
		c.Issuer = issuer
		delete(claims, "iss")
	}
	if audience, ok := claims["aud"].(string); ok {
		c.Audience = audience
		delete(claims, "aud")
	}
	if subject, ok := claims["sub"].(string); ok {
		c.Subject = subject
		delete(claims, "sub")
	}
	if issuedAt, ok := claims["iat"].(float64); ok {
		c.IssuedAt = int64(issuedAt)
		delete(claims, "iat")
	}
	if expiresAt, ok := claims["exp"].(float64); ok {
		c.ExpiresAt = int64(expiresAt)
		delete(claims, "exp")
	}
}

// Get returns a custom claim.
func (c *Claims) Get(key string) (any, bool) {
	if c.Custom == nil {
		return nil, false
	}

	value, ok := c.Custom[key]
	return value, ok
}

// GetString returns a custom string claim.
func (c *Claims) GetString(key string) string {
	if c.Custom == nil {
		return ""
	}

	value, ok := c.Custom[key].(string)
	if !ok {
		return ""
	}

	return value
}
