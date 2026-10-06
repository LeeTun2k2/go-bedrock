# Using postgresx

## Install

```sh
go get github.com/leetun2k2/go-bedrock/postgresx@latest
```

## Configure

- `URL`: PostgreSQL connection URL (e.g. `postgres://user:password@localhost:5432/mydb`). Required. Empty string is rejected. Never logged or exposed in error messages.
- `MaxConns`: Maximum number of connections in the pool. Defaults to `10`. Must be greater than zero.
- `MinConns`: Minimum number of idle connections maintained in the pool. Defaults to `2` (capped at `MaxConns`). Must not be negative or exceed `MaxConns`.
- `MaxConnLifetime`: Maximum lifetime of an individual connection before it is closed and replaced. Defaults to `1 hour`.
- `MaxConnIdleTime`: Maximum duration an idle connection can remain in the pool before being closed. Defaults to `30 minutes`.
- `HealthCheckPeriod`: Interval between background health checks of idle connections. Defaults to `1 minute`.
- `ConnectTimeout`: Timeout when establishing a single connection. Defaults to `5 seconds`.

Negative duration or count values are rejected.

## Setup

```go
ctx := context.Background()
db, err := postgresx.New(ctx, postgresx.Config{
	URL: "postgres://user:secret@localhost:5432/mydb?sslmode=disable",
})
if err != nil {
	log.Fatal(err)
}
defer db.Close()
```

## Integrate

```go
package main

import (
	"context"
	"log"

	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-bedrock/postgresx"
	"github.com/leetun2k2/go-bedrock/serverx"
)

func main() {
	ctx := context.Background()

	logger, err := logx.New(logx.LoggerConfig{Production: true})
	if err != nil {
		log.Fatal(err)
	}

	// 1. Config creation
	cfg := postgresx.Config{
		URL:      "postgres://postgres:secret@localhost:5432/mydb?sslmode=disable",
		MaxConns: 20,
		MinConns: 5,
	}

	// 2. Pool initialization (verifies connectivity with Ping)
	db, err := postgresx.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 3. Access to the underlying pgx pool for queries and transactions
	var result int
	if err := db.Pool().QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		log.Fatal(err)
	}

	// 4. Registration with serverx
	supervisor := serverx.New(serverx.Config{Name: "orders"}, logger)
	if err := supervisor.Register(db.Component()); err != nil {
		log.Fatal(err)
	}

	// 5. Run supervisor and handle graceful shutdown
	if err := supervisor.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

## Lifecycle

```mermaid
flowchart LR
    A[Config & defaults] --> B[Parse & validate]
    B --> C[Create pool]
    C --> D[Ping verification]
    D --> E[Register with serverx]
    E --> F[Graceful shutdown]
```

## Avoid

- Do not log or expose raw connection URLs or database credentials.
- Do not pass negative connection counts or timeouts.
- Do not configure `MinConns` greater than `MaxConns`.
- Do not execute database queries after the pool has closed.
