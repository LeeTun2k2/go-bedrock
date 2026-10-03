package mcpx

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// defaultAddr binds loopback only, so a default deployment is not
// reachable from outside the host. GCP deployments must set Addr
// explicitly to the interface the ingress path requires.
const defaultAddr = "127.0.0.1:7000"

const defaultPath = "/mcp"

// Config configures the mcpx Server.
//
// Addr and Path are optional: an empty value keeps the defaults
// ("127.0.0.1:7000" and "/mcp").
//
// AllowedOrigins lists the browser Origins trusted for cross-origin
// requests (e.g. "https://app.example.com"). Same-origin and non-browser
// requests are always allowed; any other Origin is rejected with 403.
//
// Stateful opts into the legacy, session-based MCP transport and requires a
// positive SessionTimeout for idle sessions to be reclaimed. The default,
// Stateful false, runs the modern stateless transport and ignores
// SessionTimeout.
type Config struct {
	Addr string `json:"addr" yaml:"addr"`
	Path string `json:"path" yaml:"path"`

	AllowedOrigins []string `json:"allowedOrigins" yaml:"allowedOrigins"`

	Stateful       bool          `json:"stateful" yaml:"stateful"`
	SessionTimeout time.Duration `json:"sessionTimeout" yaml:"sessionTimeout"`
}

func (c Config) withDefaults() Config {
	if c.Addr == "" {
		c.Addr = defaultAddr
	}

	if c.Path == "" {
		c.Path = defaultPath
	}

	return c
}

// validate reports whether c holds a usable configuration. It is called
// after withDefaults.
func (c Config) validate() error {
	if !strings.HasPrefix(c.Path, "/") {
		return fmt.Errorf("mcpx: path %q must start with \"/\"", c.Path)
	}

	if c.SessionTimeout < 0 {
		return fmt.Errorf("mcpx: session timeout must not be negative: %s", c.SessionTimeout)
	}

	if c.Stateful && c.SessionTimeout <= 0 {
		return errors.New("mcpx: stateful mode requires a positive session timeout")
	}

	return nil
}
