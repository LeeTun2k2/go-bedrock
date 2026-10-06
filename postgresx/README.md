# postgresx

`postgresx` manages a production-ready PostgreSQL connection pool backed by `pgxpool` and adapts it to the shared `serverx.Component` lifecycle.

It provides safe connection and timeout defaults, validates pool configuration, verifies database connectivity on startup, exposes the underlying `*pgxpool.Pool` for queries and transactions, and performs safe, idempotent, deadline-bounded graceful shutdown.

Use it for PostgreSQL database access in services running under `serverx`.

See [USAGE.md](USAGE.md) for installation and a complete example.
