package otelx

// Config configures the OpenTelemetry bootstrap.
type Config struct {
	ServiceName      string `json:"serviceName" yaml:"serviceName"`
	Version          string `json:"version" yaml:"version"`
	Environment      string `json:"environment" yaml:"environment"`
	Endpoint         string `json:"endpoint" yaml:"endpoint"`
	Insecure         bool   `json:"insecure" yaml:"insecure"`
	EnablePrometheus bool   `json:"enablePrometheus" yaml:"enablePrometheus"`
}
