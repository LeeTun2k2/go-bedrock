package serverx

import "time"

// Config configures the Server's own behavior.
//
// AdminAddr and ShutdownTimeout are optional: a zero value keeps the
// defaults (":4000" and 15s). A negative ShutdownTimeout is
// invalid and is replaced by the default.
type Config struct {
	Name       string `json:"name" yaml:"name"`
	Production bool   `json:"production" yaml:"production"`

	AdminAddr       string        `json:"adminAddr" yaml:"adminAddr"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout" yaml:"shutdownTimeout"`
}
