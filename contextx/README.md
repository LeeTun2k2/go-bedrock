# contextx

`contextx` keeps common service metadata in `context.Context` without repeating key names across packages.

It provides typed helpers for request, identity, tracing, messaging, deployment, and other shared string values. Generic string, integer, and boolean helpers are also available for values outside the built-in set.

Use it when metadata such as a request ID, user ID, or tenant ID must cross handler, service, and client boundaries with the same context.

See [USAGE.md](USAGE.md) for installation and a complete example.
