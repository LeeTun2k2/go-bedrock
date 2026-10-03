# serverx

`serverx` supervises the long-running parts of a Go service as one process.

It starts registered components with a shared context, stops all work when one component fails or the process receives a shutdown signal, and enforces a common shutdown deadline. A built-in admin server exposes status, Prometheus metrics, and pprof routes.

Use it to run HTTP, gRPC, MCP, and background workers under one lifecycle.

See [USAGE.md](USAGE.md) for installation and a complete example.
