package grpcx

// defaultAddr is the listen address used when Config.Addr is empty.
const defaultAddr = ":6000"

// Config configures the grpcx Server.
//
// Addr is optional: an empty value falls back to defaultAddr.
type Config struct {
	Addr string `json:"addr" yaml:"addr"`
}

func (c Config) addr() string {
	if c.Addr == "" {
		return defaultAddr
	}

	return c.Addr
}
