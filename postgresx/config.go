package postgresx

import (
	"errors"
	"fmt"
	"time"
)

const (
	defaultMaxConns          = int32(10)
	defaultMinConns          = int32(2)
	defaultMaxConnLifetime   = 1 * time.Hour
	defaultMaxConnIdleTime   = 30 * time.Minute
	defaultHealthCheckPeriod = 1 * time.Minute
	defaultConnectTimeout    = 5 * time.Second
)

// Config configures the PostgreSQL connection pool.
//
// URL is required: an empty value is rejected.
// MaxConns, MinConns, MaxConnLifetime, MaxConnIdleTime, HealthCheckPeriod,
// and ConnectTimeout are optional: zero values fall back to production defaults.
type Config struct {
	URL               string        `json:"url" yaml:"url"`
	MaxConns          int32         `json:"maxConns" yaml:"maxConns"`
	MinConns          int32         `json:"minConns" yaml:"minConns"`
	MaxConnLifetime   time.Duration `json:"maxConnLifetime" yaml:"maxConnLifetime"`
	MaxConnIdleTime   time.Duration `json:"maxConnIdleTime" yaml:"maxConnIdleTime"`
	HealthCheckPeriod time.Duration `json:"healthCheckPeriod" yaml:"healthCheckPeriod"`
	ConnectTimeout    time.Duration `json:"connectTimeout" yaml:"connectTimeout"`
}

// withDefaults returns a copy of c with zero-value settings populated with defaults.
func (c Config) withDefaults() Config {
	if c.MaxConns == 0 {
		c.MaxConns = defaultMaxConns
	}

	if c.MinConns == 0 {
		c.MinConns = defaultMinConns
		if c.MinConns > c.MaxConns {
			c.MinConns = c.MaxConns
		}
	}

	if c.MaxConnLifetime == 0 {
		c.MaxConnLifetime = defaultMaxConnLifetime
	}

	if c.MaxConnIdleTime == 0 {
		c.MaxConnIdleTime = defaultMaxConnIdleTime
	}

	if c.HealthCheckPeriod == 0 {
		c.HealthCheckPeriod = defaultHealthCheckPeriod
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}

	return c
}

// validate reports whether c holds a usable configuration.
func (c Config) validate() error {
	if c.URL == "" {
		return errors.New("postgresx: url must not be empty")
	}

	if c.MaxConns < 0 {
		return fmt.Errorf("postgresx: max connections must not be negative: %d", c.MaxConns)
	}

	if c.MaxConns == 0 {
		return errors.New("postgresx: max connections must be greater than zero")
	}

	if c.MinConns < 0 {
		return fmt.Errorf("postgresx: min connections must not be negative: %d", c.MinConns)
	}

	if c.MinConns > c.MaxConns {
		return fmt.Errorf("postgresx: min connections (%d) must not exceed max connections (%d)", c.MinConns, c.MaxConns)
	}

	if c.MaxConnLifetime < 0 {
		return fmt.Errorf("postgresx: max connection lifetime must not be negative: %s", c.MaxConnLifetime)
	}

	if c.MaxConnIdleTime < 0 {
		return fmt.Errorf("postgresx: max idle time must not be negative: %s", c.MaxConnIdleTime)
	}

	if c.HealthCheckPeriod < 0 {
		return fmt.Errorf("postgresx: health-check period must not be negative: %s", c.HealthCheckPeriod)
	}

	if c.ConnectTimeout < 0 {
		return fmt.Errorf("postgresx: connect timeout must not be negative: %s", c.ConnectTimeout)
	}

	return nil
}

// String returns a safe string representation of Config with the connection URL redacted.
func (c Config) String() string {
	return fmt.Sprintf("postgresx.Config{URL: %q, MaxConns: %d, MinConns: %d, MaxConnLifetime: %s, MaxConnIdleTime: %s, HealthCheckPeriod: %s, ConnectTimeout: %s}",
		"[redacted]",
		c.MaxConns,
		c.MinConns,
		c.MaxConnLifetime,
		c.MaxConnIdleTime,
		c.HealthCheckPeriod,
		c.ConnectTimeout,
	)
}
