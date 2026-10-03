// Package serverx provides a long-running process context, pre-configured
// with a logger, that hosts one or more Components (HTTP, gRPC, MCP,
// workers, ...) alongside a built-in admin server exposing health checks,
// Prometheus metrics, and pprof profiling.
package serverx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LeeTun2k2/go-bedrock/logx"
)

const (
	defaultAdminAddr       = ":9000"
	defaultShutdownTimeout = 15 * time.Second
)

// ErrShutdownTimeout is returned by Run, alone or joined with a component's
// failure, when shutdown does not complete before the configured
// ShutdownTimeout.
var ErrShutdownTimeout = errors.New("serverx: shutdown deadline exceeded")

// Server runs the admin endpoints and any Components registered into it.
// All Components share the Server's context: if one stops or fails, that
// context is canceled and every other Component, along with the admin
// server, is stopped too.
type Server struct {
	cfg    Config
	logger *logx.Logger
	admin  *http.Server
	cancel context.CancelFunc

	mu         sync.Mutex
	started    bool
	components []Component
	healthy    bool
	detail     string
	err        error

	wg sync.WaitGroup
}

// New creates a Server with the given configuration and logger. Register
// Components on the returned Server, then call Run to start it.
func New(cfg Config, logger *logx.Logger) *Server {
	if cfg.AdminAddr == "" {
		cfg.AdminAddr = defaultAdminAddr
	}

	switch {
	case cfg.ShutdownTimeout == 0:
		cfg.ShutdownTimeout = defaultShutdownTimeout
	case cfg.ShutdownTimeout < 0:
		logger.Error(context.Background(), "negative shutdown timeout, using default",
			"configured", cfg.ShutdownTimeout, "default", defaultShutdownTimeout)
		cfg.ShutdownTimeout = defaultShutdownTimeout
	}

	s := &Server{
		cfg:    cfg,
		logger: logger,
	}
	s.admin = &http.Server{
		Addr:    cfg.AdminAddr,
		Handler: newAdminMux(s),
	}

	return s
}

// Register adds c to the Server. It must be called before Run; a call made
// once Run has started returns an error and does not retain c. Start is
// required and rejected if nil; a nil Stop is treated as a no-op.
func (s *Server) Register(c Component) error {
	if c.Start == nil {
		return fmt.Errorf("serverx: component %s: Start must not be nil", c.Name)
	}

	if c.Stop == nil {
		c.Stop = func(context.Context) error { return nil }
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("serverx: component %s: Register called after Run has started", c.Name)
	}

	s.components = append(s.components, c)

	return nil
}

// Run starts the admin server and every registered Component, then blocks
// until ctx is canceled, an interrupt/terminate signal is received, or a
// Component fails. It then stops all work concurrently and returns once
// that work finishes or the configured ShutdownTimeout elapses, whichever
// comes first. Run must not be called more than once.
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("serverx: Run has already been called")
	}
	s.started = true
	s.healthy = true
	components := s.components
	s.mu.Unlock()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.wg.Add(1)
	go s.runAdmin(runCtx)

	for _, c := range components {
		s.wg.Add(1)
		go s.runComponent(runCtx, c)
	}

	<-runCtx.Done()

	s.markUnhealthy("server is shutting down")

	timedOut := s.shutdown()

	s.mu.Lock()
	rootErr := s.err
	s.mu.Unlock()

	if timedOut {
		return errors.Join(rootErr, ErrShutdownTimeout)
	}

	return rootErr
}

func (s *Server) runAdmin(ctx context.Context) {
	defer s.wg.Done()

	s.logger.Info(ctx, "admin server starting", "addr", s.cfg.AdminAddr)

	if err := s.admin.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.fail(fmt.Errorf("admin server: %w", err))
	}
}

func (s *Server) runComponent(ctx context.Context, c Component) {
	defer s.wg.Done()

	s.logger.Info(ctx, "component starting", "component", c.Name)

	err := c.Start(ctx)
	if ctx.Err() != nil {
		// Shutdown is already underway; this return does not start
		// another failure cascade.
		return
	}

	if err == nil {
		err = fmt.Errorf("component %s: exited unexpectedly", c.Name)
	} else {
		err = fmt.Errorf("component %s: %w", c.Name, err)
	}

	s.fail(err)
}

// fail records err as the root cause of shutdown, if none is already
// recorded, and cancels the server context.
func (s *Server) fail(err error) {
	s.mu.Lock()
	first := s.healthy
	if first {
		s.healthy = false
		s.detail = err.Error()
		s.err = err
	}
	s.mu.Unlock()

	if first {
		s.logger.Error(context.Background(), "component failed, shutting down", "error", err)
	}

	s.cancel()
}

// markUnhealthy records detail as the status reason, if no failure has
// already recorded one.
func (s *Server) markUnhealthy(detail string) {
	s.mu.Lock()
	if s.healthy {
		s.healthy = false
		s.detail = detail
	}
	s.mu.Unlock()
}

// shutdown stops the admin server and every Component concurrently, bounded
// by the configured ShutdownTimeout. It reports whether that deadline was
// reached before all work finished.
func (s *Server) shutdown() bool {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	go func() {
		if err := s.admin.Shutdown(shutdownCtx); err != nil {
			s.logger.Error(shutdownCtx, "admin server shutdown failed", "error", err)
		}
	}()

	s.mu.Lock()
	components := s.components
	s.mu.Unlock()

	for _, c := range components {
		go func(c Component) {
			s.logger.Info(shutdownCtx, "component stopping", "component", c.Name)

			if err := c.Stop(shutdownCtx); err != nil {
				s.logger.Error(shutdownCtx, "component stop failed", "component", c.Name, "error", err)
			}
		}(c)
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return false
	case <-shutdownCtx.Done():
		return true
	}
}
