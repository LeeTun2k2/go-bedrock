# restapix

`restapix` adapts a Gin router and hardened HTTP server to the shared `serverx.Component` lifecycle.

It provides safe header and idle timeout defaults, explicit trusted-proxy handling, panic recovery through `logx`, route registration through the underlying Gin engine, and graceful shutdown.

Use it for REST APIs that run under `serverx` and need clear network defaults.

See [USAGE.md](USAGE.md) for installation and a complete example.
