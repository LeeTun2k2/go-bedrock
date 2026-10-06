package redisx

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr         = "127.0.0.1:6379"
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 3 * time.Second
	defaultWriteTimeout = 3 * time.Second
	defaultPoolSize     = 10
	defaultMinIdleConns = 2
	defaultMaxIdleTime  = 30 * time.Minute
	defaultMaxConnAge   = 1 * time.Hour
)

// Config configures the standalone Redis client connection and pool.
//
// Zero values for Addr, DialTimeout, ReadTimeout, WriteTimeout, PoolSize,
// MinIdleConns, MaxIdleTime, and MaxConnAge fall back to production defaults.
// Username and Password default to empty (unauthenticated). DB defaults to 0.
// TLSEnabled defaults to false.
type Config struct {
	Addr          string        `json:"addr" yaml:"addr"`
	Username      string        `json:"username" yaml:"username"`
	Password      string        `json:"password" yaml:"password"`
	DB            int           `json:"db" yaml:"db"`
	TLSEnabled    bool          `json:"tlsEnabled" yaml:"tlsEnabled"`
	TLSServerName string        `json:"tlsServerName" yaml:"tlsServerName"`
	DialTimeout   time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ReadTimeout   time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout  time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
	PoolSize      int           `json:"poolSize" yaml:"poolSize"`
	MinIdleConns  int           `json:"minIdleConns" yaml:"minIdleConns"`
	MaxIdleTime   time.Duration `json:"maxIdleTime" yaml:"maxIdleTime"`
	MaxConnAge    time.Duration `json:"maxConnAge" yaml:"maxConnAge"`
}

// withDefaults returns a copy of c with zero-value settings populated with defaults.
func (c Config) withDefaults() Config {
	if c.Addr == "" {
		c.Addr = defaultAddr
	}

	if c.DialTimeout == 0 {
		c.DialTimeout = defaultDialTimeout
	}

	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultReadTimeout
	}

	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultWriteTimeout
	}

	if c.PoolSize == 0 {
		c.PoolSize = defaultPoolSize
	}

	if c.MinIdleConns == 0 {
		c.MinIdleConns = defaultMinIdleConns
		if c.MinIdleConns > c.PoolSize {
			c.MinIdleConns = c.PoolSize
		}
	}

	if c.MaxIdleTime == 0 {
		c.MaxIdleTime = defaultMaxIdleTime
	}

	if c.MaxConnAge == 0 {
		c.MaxConnAge = defaultMaxConnAge
	}

	return c
}

// validate reports whether c holds a usable configuration.
func (c Config) validate() error {
	if err := validateAddr(c.Addr); err != nil {
		return err
	}

	if c.DB < 0 {
		return fmt.Errorf("redisx: database number must not be negative: %d", c.DB)
	}

	if !c.TLSEnabled {
		if c.TLSServerName != "" {
			return errors.New("redisx: tls server name cannot be set when tls is disabled")
		}
	}

	if c.DialTimeout < 0 {
		return fmt.Errorf("redisx: dial timeout must not be negative: %s", c.DialTimeout)
	}

	if c.ReadTimeout < 0 {
		return fmt.Errorf("redisx: read timeout must not be negative: %s", c.ReadTimeout)
	}

	if c.WriteTimeout < 0 {
		return fmt.Errorf("redisx: write timeout must not be negative: %s", c.WriteTimeout)
	}

	if c.MaxIdleTime < 0 {
		return fmt.Errorf("redisx: max idle time must not be negative: %s", c.MaxIdleTime)
	}

	if c.MaxConnAge < 0 {
		return fmt.Errorf("redisx: max connection age must not be negative: %s", c.MaxConnAge)
	}

	if c.PoolSize < 0 {
		return fmt.Errorf("redisx: pool size must not be negative: %d", c.PoolSize)
	}

	if c.PoolSize == 0 {
		return errors.New("redisx: pool size must be greater than zero")
	}

	if c.MinIdleConns < 0 {
		return fmt.Errorf("redisx: min idle connections must not be negative: %d", c.MinIdleConns)
	}

	if c.MinIdleConns > c.PoolSize {
		return fmt.Errorf("redisx: min idle connections (%d) must not exceed pool size (%d)", c.MinIdleConns, c.PoolSize)
	}

	return nil
}

// validateAddr verifies that addr is a non-empty host:port string with a valid port number.
func validateAddr(addr string) error {
	if strings.TrimSpace(addr) == "" {
		return errors.New("redisx: address must not be empty")
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("redisx: invalid address %q: %w", addr, err)
	}

	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("redisx: address host must not be empty: %q", addr)
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("redisx: invalid port %q in address: %w", port, err)
	}

	if p <= 0 {
		return fmt.Errorf("redisx: port must be greater than zero: %d", p)
	}

	if p > 65535 {
		return fmt.Errorf("redisx: port out of range: %d", p)
	}

	return nil
}

// String returns a safe string representation of Config with the password redacted.
func (c Config) String() string {
	passwordStr := ""
	if c.Password != "" {
		passwordStr = "[redacted]"
	}

	return fmt.Sprintf("redisx.Config{Addr: %q, Username: %q, Password: %q, DB: %d, TLSEnabled: %t, TLSServerName: %q, DialTimeout: %s, ReadTimeout: %s, WriteTimeout: %s, PoolSize: %d, MinIdleConns: %d, MaxIdleTime: %s, MaxConnAge: %s}",
		c.Addr,
		c.Username,
		passwordStr,
		c.DB,
		c.TLSEnabled,
		c.TLSServerName,
		c.DialTimeout,
		c.ReadTimeout,
		c.WriteTimeout,
		c.PoolSize,
		c.MinIdleConns,
		c.MaxIdleTime,
		c.MaxConnAge,
	)
}
