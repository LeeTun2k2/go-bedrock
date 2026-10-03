# Using restapix

## Install

```sh
go get github.com/LeeTun2k2/go-bedrock/restapix@latest
```

## Configure

- `Addr` defaults to `:5000`.
- `ReadHeaderTimeout` defaults to 5 seconds.
- `IdleTimeout` defaults to 60 seconds.
- `ReadTimeout` and `WriteTimeout` are disabled by default. Set them only when whole-request deadlines are safe.
- An empty `TrustedProxies` list trusts no proxy. Add only known proxy IPs or CIDRs.

Negative timeout values are rejected.

## Integrate

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/LeeTun2k2/go-bedrock/logx"
	"github.com/LeeTun2k2/go-bedrock/restapix"
	"github.com/LeeTun2k2/go-bedrock/serverx"
	"github.com/gin-gonic/gin"
)

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func main() {
	logger, err := logx.New(logx.LoggerConfig{Production: true})
	if err != nil {
		log.Fatal(err)
	}

	api, err := restapix.New(restapix.Config{
		Addr:           ":5000",
		TrustedProxies: []string{"10.0.0.0/8"},
	}, logger)
	if err != nil {
		log.Fatal(err)
	}
	api.Engine().GET("/health", healthHandler)

	supervisor := serverx.New(serverx.Config{Name: "orders"}, logger)
	if err := supervisor.Register(api.Component()); err != nil {
		log.Fatal(err)
	}

	if err := supervisor.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

## Lifecycle

```mermaid
flowchart LR
    A[Create] --> B[Add middleware and routes]
    B --> C[Register component]
    C --> D[Run supervisor]
    D --> E[Graceful shutdown]
```

## Avoid

- Do not add routes or global middleware after the server starts.
- Do not trust broad proxy ranges unless the network path enforces them.
- Do not enable whole-request timeouts for streaming or long transfers without checking their effect.
- Do not pass a nil logger.
