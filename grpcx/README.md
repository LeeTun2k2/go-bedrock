# grpcx

`grpcx` adapts a standard gRPC server to the shared `serverx.Component` lifecycle.

It creates the underlying `grpc.Server`, exposes it for service registration, listens on a configured TCP address, and performs deadline-bound graceful shutdown with forced fallback.

Use it when a process managed by `serverx` serves one or more generated gRPC services.

See [USAGE.md](USAGE.md) for installation and a complete example.
