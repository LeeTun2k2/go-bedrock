// Package mcpx adapts an MCP server, served over Streamable HTTP, into a
// serverx.Component.
package mcpx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/LeeTun2k2/go-bedrock/logx"
	"github.com/LeeTun2k2/go-bedrock/serverx"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Option configures optional aspects of the MCP server constructed by New.
type Option func(*options)

type options struct {
	implementation *mcp.Implementation
	serverOptions  *mcp.ServerOptions
	streamableOpts *mcp.StreamableHTTPOptions
	middleware     []func(http.Handler) http.Handler
}

// WithImplementation sets the MCP server's Implementation (name/version)
// metadata. Defaults to &mcp.Implementation{Name: "mcpx"}. Passing nil is a
// configuration error: New returns it instead of constructing a server with
// no identity.
func WithImplementation(impl *mcp.Implementation) Option {
	return func(o *options) { o.implementation = impl }
}

// WithServerOptions sets the options passed to mcp.NewServer.
func WithServerOptions(so *mcp.ServerOptions) Option {
	return func(o *options) { o.serverOptions = so }
}

// WithStreamableHTTPOptions sets the options passed to
// mcp.NewStreamableHTTPHandler. Stateless, SessionTimeout, and
// PropagateRequestCancellation are governed by Config and are overwritten
// by New; set other fields (EventStore, JSONResponse, MaxRequestBodyBytes,
// ...) here.
func WithStreamableHTTPOptions(so *mcp.StreamableHTTPOptions) Option {
	return func(o *options) { o.streamableOpts = so }
}

// WithMiddleware wraps the complete MCP HTTP handler with mw, applied
// outermost-first, so every mcpx request passes through it before origin
// protection and the MCP handler itself. Use it for auth, rate limiting,
// tracing, or logging.
func WithMiddleware(mw ...func(http.Handler) http.Handler) Option {
	return func(o *options) { o.middleware = append(o.middleware, mw...) }
}

// Server adapts an MCP server, served over Streamable HTTP, into a
// serverx.Component.
type Server struct {
	cfg        Config
	logger     *logx.Logger
	mcpServer  *mcp.Server
	httpServer *http.Server
}

// New constructs the underlying MCP server and internal HTTP server. It does
// not start listening yet. Register tools, resources, and prompts on the
// server returned by MCPServer before calling Start. New returns an error,
// instead of a Server, for an invalid path, a nil implementation, or an
// unsafe stateful configuration (stateful mode without a positive
// SessionTimeout).
func New(cfg Config, logger *logx.Logger, opts ...Option) (*Server, error) {
	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	o := &options{
		implementation: &mcp.Implementation{Name: "mcpx"},
	}
	for _, opt := range opts {
		opt(o)
	}

	if o.implementation == nil {
		return nil, errors.New("mcpx: implementation must not be nil")
	}

	protection := http.NewCrossOriginProtection()
	for _, origin := range cfg.AllowedOrigins {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			return nil, fmt.Errorf("mcpx: invalid allowed origin %q: %w", origin, err)
		}
	}

	mcpServer := mcp.NewServer(o.implementation, o.serverOptions)

	streamableOpts := mcp.StreamableHTTPOptions{}
	if o.streamableOpts != nil {
		streamableOpts = *o.streamableOpts
	}
	streamableOpts.Stateless = !cfg.Stateful
	streamableOpts.SessionTimeout = cfg.SessionTimeout
	streamableOpts.PropagateRequestCancellation = true

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, &streamableOpts)

	var wrapped http.Handler = protection.Handler(handler)
	for i := len(o.middleware) - 1; i >= 0; i-- {
		wrapped = o.middleware[i](wrapped)
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, wrapped)

	return &Server{
		cfg:       cfg,
		logger:    logger,
		mcpServer: mcpServer,
		httpServer: &http.Server{
			Addr:              cfg.Addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

// MCPServer exposes the underlying MCP server so callers can register tools,
// resources, and prompts before Start.
func (s *Server) MCPServer() *mcp.Server {
	return s.mcpServer
}

// Component returns a serverx.Component adapting Start/Stop.
func (s *Server) Component() serverx.Component {
	return serverx.Component{
		Name:  "mcpx",
		Start: s.Start,
		Stop:  s.Stop,
	}
}

// Start binds Config.Addr and serves the Streamable HTTP handler. It blocks
// until ctx is done or the server fails, returning the real error on a
// listen failure.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		s.logger.Error(ctx, "mcpx: failed to listen", "addr", s.cfg.Addr, "error", err)
		return fmt.Errorf("mcpx: listen: %w", err)
	}

	s.logger.Info(ctx, "mcpx server starting", "addr", s.cfg.Addr, "path", s.cfg.Path)

	if err := s.httpServer.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error(ctx, "mcpx: serve failed", "addr", s.cfg.Addr, "error", err)
		return fmt.Errorf("mcpx: serve: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server, bounded by ctx's deadline.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(ctx, "mcpx: shutdown failed", "error", err)
		return fmt.Errorf("mcpx: shutdown: %w", err)
	}

	return nil
}
