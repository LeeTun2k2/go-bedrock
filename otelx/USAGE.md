# Using otelx

## Install

```sh
go get github.com/LeeTun2k2/go-bedrock/otelx@latest
```

## Configure

- `ServiceName`, `Version`, and `Environment` become resource attributes.
- `Endpoint` is the OTLP gRPC collector address. An empty value disables trace export.
- `Insecure` enables plaintext transport to the collector. Keep it `false` when TLS is required.
- `EnablePrometheus` installs a Prometheus meter provider. Expose the registry with a compatible HTTP handler, such as the `serverx` admin server.

## Integrate

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/LeeTun2k2/go-bedrock/otelx"
)

func main() {
	shutdown, err := otelx.Init(context.Background(), otelx.Config{
		ServiceName:      "orders",
		Version:          "1.0.0",
		Environment:      "production",
		Endpoint:         "otel-collector:4317",
		EnablePrometheus: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := shutdown(shutdownCtx); err != nil {
		log.Printf("stop telemetry: %v", err)
	}
}
```

## Lifecycle

Call `Init` exactly once during process startup. Keep its shutdown function and call it once with a bounded context after application work stops.

```mermaid
flowchart LR
    A[Process starts] --> B[Init once]
    B --> C[Run service]
    C --> D[Stop service]
    D --> E[Shutdown with timeout]
```

The returned shutdown function is idempotent.

## Avoid

- Do not call `Init` more than once in a process.
- Do not use `Insecure` across an untrusted network.
- Do not call shutdown with an unbounded context.
- Do not assume an empty endpoint exports traces; it creates a provider without an exporter.
