// Package otelx bootstraps OpenTelemetry: an OTLP gRPC trace pipeline, a
// Prometheus metrics pipeline, and the W3C propagator. Callers use the
// official go.opentelemetry.io/otel API directly; this package only wires
// global providers.
package otelx

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"google.golang.org/grpc/credentials/insecure"
)

// ShutdownFunc flushes and closes telemetry exporters. It is idempotent and
// respects the deadline of the ctx passed to it; callers must supply a
// bounded timeout context.
type ShutdownFunc func(ctx context.Context) error

// Init registers global TracerProvider, MeterProvider, and W3C propagator
// for the process. It must be called exactly once per process.
//
// If cfg.Endpoint is empty, trace export is a no-op; Prometheus metrics
// remain fully functional when cfg.EnablePrometheus is true.
func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	var stopFuncs []func(ctx context.Context) error

	tracerProvider, stopTraces, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return nil, err
	}
	stopFuncs = append(stopFuncs, stopTraces)
	otel.SetTracerProvider(tracerProvider)

	if cfg.EnablePrometheus {
		meterProvider, stopMetrics, err := newMeterProvider(res)
		if err != nil {
			return nil, err
		}
		stopFuncs = append(stopFuncs, stopMetrics)
		otel.SetMeterProvider(meterProvider)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return newShutdownFunc(stopFuncs), nil
}

func newTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, func(ctx context.Context) error, error) {
	if cfg.Endpoint == "" {
		provider := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		return provider, provider.Shutdown, nil
	}

	exporterOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	return provider, provider.Shutdown, nil
}

func newMeterProvider(res *resource.Resource) (*sdkmetric.MeterProvider, func(ctx context.Context) error, error) {
	reader, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	return provider, provider.Shutdown, nil
}

func newShutdownFunc(stopFuncs []func(ctx context.Context) error) ShutdownFunc {
	var once sync.Once
	var shutdownErr error

	return func(ctx context.Context) error {
		once.Do(func() {
			var errs []error
			for _, stop := range stopFuncs {
				if err := stop(ctx); err != nil {
					errs = append(errs, err)
				}
			}
			shutdownErr = errors.Join(errs...)
		})
		return shutdownErr
	}
}
