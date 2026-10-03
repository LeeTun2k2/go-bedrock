package logx

import (
	"context"

	"github.com/LeeTun2k2/go-bedrock/contextx"
	"go.opentelemetry.io/otel/trace"
)

func ContextFields(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}

	fields := make([]any, 0, 8)

	if v, ok := ctx.Value(contextx.KeyRequestID).(string); ok && v != "" {
		fields = append(fields, "request_id", v)
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		fields = append(
			fields,
			"trace_id", spanContext.TraceID().String(),
			"span_id", spanContext.SpanID().String(),
		)

		if spanContext.IsSampled() {
			fields = append(fields, "trace_sampled", true)
		}
	}

	return fields
}
