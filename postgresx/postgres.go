// Package postgresx provides a connection pool for PostgreSQL backed by pgxpool,
// adapted to the serverx.Component lifecycle.
package postgresx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leetun2k2/go-bedrock/serverx"
)

var cannotParseRegex = regexp.MustCompile("cannot parse `[^`]+`:\\s*")

// DB wraps a *pgxpool.Pool and adapts it to serverx.Component.
type DB struct {
	pool      *pgxpool.Pool
	closed    chan struct{}
	closeOnce sync.Once
}

// New constructs and connects a PostgreSQL connection pool using pgxpool.
// It verifies connectivity with Ping before returning. New returns an error
// if the configuration is invalid or if connecting/pinging fails.
func New(ctx context.Context, cfg Config) (*DB, error) {
	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("postgresx: parse config: %w", sanitizeError(err, cfg.URL))
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	if poolCfg.ConnConfig != nil {
		poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgresx: create pool: %w", sanitizeError(err, cfg.URL))
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgresx: ping: %w", sanitizeError(err, cfg.URL))
	}

	return &DB{
		pool:   pool,
		closed: make(chan struct{}),
	}, nil
}

// Pool exposes the underlying *pgxpool.Pool for queries and transactions.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Ping checks whether the PostgreSQL database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Close closes all connections in the pool. It is safe and idempotent.
func (db *DB) Close() {
	db.closeOnce.Do(func() {
		close(db.closed)
		db.pool.Close()
	})
}

// Component adapts DB to a serverx.Component for graceful shutdown.
func (db *DB) Component() serverx.Component {
	return serverx.Component{
		Name:  "postgresx",
		Start: db.Start,
		Stop:  db.Stop,
	}
}

// Start blocks until ctx is canceled or the pool is closed.
func (db *DB) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-db.closed:
		return nil
	}
}

// Stop gracefully closes the pool, bounded by ctx's deadline.
func (db *DB) Stop(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		db.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// sanitizeError removes connection URLs, credentials, and passwords from error messages.
func sanitizeError(err error, rawURL string) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	// Strip backticked URL from pgconn "cannot parse `...`: reason"
	msg = cannotParseRegex.ReplaceAllString(msg, "invalid connection url: ")

	// Replace the full raw connection URL if present
	if rawURL != "" {
		msg = strings.ReplaceAll(msg, rawURL, "[redacted]")
	}

	// Redact parsed password and credentials if present
	if u, parseErr := url.Parse(rawURL); parseErr == nil {
		if u.User != nil {
			pass, ok := u.User.Password()
			if ok {
				if pass != "" {
					msg = strings.ReplaceAll(msg, pass, "[redacted]")
				}
			}
		}

		red := u.Redacted()
		if red != "" {
			msg = strings.ReplaceAll(msg, red, "[redacted]")
		}
	}

	return errors.New(msg)
}
