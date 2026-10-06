// Package redisx provides a production-ready standalone Redis client backed by go-redis/v9,
// adapted to the serverx.Component lifecycle.
package redisx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/leetun2k2/go-bedrock/serverx"
	"github.com/redis/go-redis/v9"
)

var credURLRegex = regexp.MustCompile(`rediss?://[^@\s]+@`)

// Client wraps a *redis.Client and adapts it to serverx.Component.
//
// redisx exposes the concrete *redis.Client rather than the redis.UniversalClient
// interface because redisx targets a standalone single-node Redis deployment.
// *redis.Client provides the smallest, most direct API surface needed for standalone
// operations without multi-node routing overhead or unnecessary interface abstractions.
type Client struct {
	client    *redis.Client
	closed    chan struct{}
	closeOnce sync.Once
	closeErr  error
}

// New constructs and connects a standalone Redis client using go-redis/v9.
// It verifies connectivity with Ping before returning. New returns an error
// if the configuration is invalid or if connecting/pinging fails. If Ping fails,
// any partially created client resources are cleanly closed.
func New(ctx context.Context, cfg Config) (*Client, error) {
	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	opts := &redis.Options{
		Addr:            cfg.Addr,
		Username:        cfg.Username,
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxIdleTime: cfg.MaxIdleTime,
		ConnMaxLifetime: cfg.MaxConnAge,
	}

	if cfg.TLSEnabled {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if cfg.TLSServerName != "" {
			tlsConfig.ServerName = cfg.TLSServerName
		} else {
			host, _, _ := net.SplitHostPort(cfg.Addr)
			if net.ParseIP(host) == nil {
				tlsConfig.ServerName = host
			}
		}
		opts.TLSConfig = tlsConfig
	}

	rdb := redis.NewClient(opts)

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redisx: ping: %w", sanitizeError(err, cfg.Password))
	}

	return &Client{
		client: rdb,
		closed: make(chan struct{}),
	}, nil
}

// Client exposes the underlying *redis.Client for Redis commands and pipelines.
func (c *Client) Client() *redis.Client {
	return c.client
}

// Ping checks whether the Redis server is reachable.
func (c *Client) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redisx: ping: %w", sanitizeError(err, c.client.Options().Password))
	}
	return nil
}

// Close closes the underlying Redis client connection pool. It is safe and idempotent.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.closeErr = c.client.Close()
	})
	return c.closeErr
}

// Component adapts Client to a serverx.Component for graceful shutdown.
func (c *Client) Component() serverx.Component {
	return serverx.Component{
		Name:  "redisx",
		Start: c.Start,
		Stop:  c.Stop,
	}
}

// Start blocks until ctx is canceled or the client is closed.
func (c *Client) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-c.closed:
		return nil
	}
}

// Stop gracefully closes the client, bounded by ctx's deadline.
func (c *Client) Stop(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		_ = c.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// sanitizeError removes credentials and sensitive passwords from error messages.
func sanitizeError(err error, password string) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	if password != "" {
		msg = strings.ReplaceAll(msg, password, "[redacted]")
	}

	msg = credURLRegex.ReplaceAllString(msg, "[redacted]@")

	return errors.New(msg)
}
