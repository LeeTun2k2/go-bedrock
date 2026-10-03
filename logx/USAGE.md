# Using logx

## Install

```sh
go get github.com/LeeTun2k2/go-bedrock/logx@latest
```

## Configure

Set `LoggerConfig.Production` to `true` for Zap's production format. The zero value uses Zap's development logger.

`ContextFields` adds a non-empty `contextx.KeyRequestID`. It also adds valid OpenTelemetry trace and span IDs, plus `trace_sampled` for sampled spans.

## Integrate

```go
package main

import (
	"context"
	"log"

	"github.com/LeeTun2k2/go-bedrock/contextx"
	"github.com/LeeTun2k2/go-bedrock/logx"
)

func main() {
	logger, err := logx.New(logx.LoggerConfig{Production: true})
	if err != nil {
		log.Fatal(err)
	}

	ctx := contextx.SetRequestID(context.Background(), "req-123")
	logger.Info(ctx, "order accepted", "order_id", "order-456")
}
```

## Lifecycle

Create one logger during process setup and reuse it. Pass the active request or task context to each log method.

## Avoid

- Do not pass secrets as fields.
- Do not pass an odd number of key-value arguments.
- Do not use `Fatal` in request code; it terminates the process.
- Do not use `Panic` for normal error handling.
