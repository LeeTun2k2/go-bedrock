# redisx

`redisx` manages a production-ready standalone Redis client backed by `go-redis/v9` and adapts it to the shared `serverx.Component` lifecycle.

It provides safe connection, timeout, and pool defaults, validates configuration, verifies Redis connectivity on startup with `Ping`, exposes the underlying `*redis.Client` for Redis commands and pipelines, and performs safe, idempotent, deadline-bounded graceful shutdown.

`redisx` targets single-node Redis deployments and exposes the concrete `*redis.Client` rather than `redis.UniversalClient` to keep the API surface minimal, direct, and free of multi-node routing overhead.

Use it for Redis access in services running under `serverx`.

See [USAGE.md](USAGE.md) for installation and a complete example.
