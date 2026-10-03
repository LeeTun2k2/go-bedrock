# otelx

`otelx` sets up process-wide OpenTelemetry tracing, metrics, and context propagation.

It installs an OTLP gRPC trace provider, an optional Prometheus meter provider, and W3C Trace Context plus Baggage propagation. Application code continues to use the standard OpenTelemetry API.

Use it when a service needs one consistent telemetry bootstrap and bounded exporter shutdown.

See [USAGE.md](USAGE.md) for installation and a complete example.
