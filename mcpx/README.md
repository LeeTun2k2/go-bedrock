# mcpx

`mcpx` serves an MCP server over Streamable HTTP under the shared `serverx.Component` lifecycle.

It provides safe loopback defaults, cross-origin protection, middleware wrapping, stateless or bounded stateful sessions, request cancellation propagation, and graceful HTTP shutdown.

Use it to expose MCP tools, resources, or prompts from a Go service managed by `serverx`.

See [USAGE.md](USAGE.md) for installation and a complete example.
