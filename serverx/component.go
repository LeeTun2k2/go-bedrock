package serverx

import "context"

// Component is a long-running unit of work run by the Server, such as an
// HTTP server, a gRPC server, an MCP server, or a background worker.
type Component struct {
	// Name identifies the component in logs.
	Name string

	// Start runs the component and blocks until ctx is canceled or the
	// component fails. It must return nil on clean shutdown. Start is
	// required; Register rejects a Component with a nil Start.
	Start func(ctx context.Context) error

	// Stop gracefully stops the component. ctx carries the shutdown
	// deadline enforced by the Server. Stop is optional; a nil Stop is
	// treated as a no-op.
	Stop func(ctx context.Context) error
}
