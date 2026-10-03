# Using mcpx

## Install

```sh
go get github.com/LeeTun2k2/go-bedrock/mcpx@latest
```

## Configure

- `Addr` defaults to `127.0.0.1:7000`. Set it explicitly when the service must accept remote traffic.
- `Path` defaults to `/mcp` and must start with `/`.
- `AllowedOrigins` contains exact browser origins allowed by cross-origin protection.
- Stateless mode is the default.
- Stateful mode requires a positive `SessionTimeout` so idle sessions are reclaimed.

Use `WithImplementation` for server identity, `WithServerOptions` for MCP behavior, `WithStreamableHTTPOptions` for other transport options, and `WithMiddleware` for auth, rate limits, tracing, or logging. Configuration always controls stateless mode, session timeout, and request cancellation propagation.

## Integrate

```go
package main

import (
	"context"
	"log"

	"github.com/LeeTun2k2/go-bedrock/logx"
	"github.com/LeeTun2k2/go-bedrock/mcpx"
	"github.com/LeeTun2k2/go-bedrock/serverx"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type echoInput struct {
	Message string `json:"message"`
}

type echoOutput struct {
	Message string `json:"message"`
}

func echoTool(
	_ context.Context,
	_ *mcp.CallToolRequest,
	input echoInput,
) (*mcp.CallToolResult, echoOutput, error) {
	return nil, echoOutput{Message: input.Message}, nil
}

func main() {
	logger, err := logx.New(logx.LoggerConfig{Production: true})
	if err != nil {
		log.Fatal(err)
	}

	mcpServer, err := mcpx.New(mcpx.Config{Addr: ":7000"}, logger)
	if err != nil {
		log.Fatal(err)
	}
	mcp.AddTool(mcpServer.MCPServer(), &mcp.Tool{
		Name:        "echo",
		Description: "Return the supplied message.",
	}, echoTool)

	supervisor := serverx.New(serverx.Config{Name: "tools"}, logger)
	if err := supervisor.Register(mcpServer.Component()); err != nil {
		log.Fatal(err)
	}

	if err := supervisor.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

For stateful mode, set both values:

```go
mcpx.Config{
	Stateful:       true,
	SessionTimeout: 30 * time.Minute,
}
```

## Lifecycle

```mermaid
flowchart LR
    A[Create] --> B[Register MCP features]
    B --> C[Register component]
    C --> D[Run supervisor]
    D --> E[Graceful shutdown]
```

## Avoid

- Do not use stateful mode without a positive session timeout.
- Do not expose the default loopback address when remote access is required.
- Do not allow untrusted browser origins.
- Do not register tools, resources, or prompts after the server starts.
- Do not pass a nil implementation or logger.
