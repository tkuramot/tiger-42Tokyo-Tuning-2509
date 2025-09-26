// internal/telemetry/telemetry.go
package telemetry

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func enabled() bool {
	if strings.EqualFold(os.Getenv("TRACE_ENABLED"), "false") {
		return false
	}
	if strings.EqualFold(os.Getenv("TRACE_ENABLED"), "true") {
		return true
	}
	if os.Getenv("JAEGER_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		return true
	}
	return false
}

func samplerFromEnv() sdktrace.Sampler {
	if r := os.Getenv("TRACE_SAMPLE_RATIO"); r != "" {
		if v, err := strconv.ParseFloat(r, 64); err == nil {
			return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(v))
		}
	}
	switch strings.ToLower(os.Getenv("OTEL_TRACES_SAMPLER")) {
	case "always_off":
		return sdktrace.NeverSample()
	case "always_on":
		return sdktrace.AlwaysSample()
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.01))
}

func resourceFromEnv() *resource.Resource {
	svc := os.Getenv("SERVICE_NAME")
	if svc == "" {
		svc = "backend"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
		if env == "" {
			env = "local"
		}
	}
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(svc),
			attribute.String("deployment.environment", env),
		),
	)
	return r
}

func Init(ctx context.Context) (func(context.Context) error, error) {
	if !enabled() {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))
		return func(context.Context) error { return nil }, nil
	}

	var (
		exp sdktrace.SpanExporter
		err error
	)
	if ep := os.Getenv("JAEGER_ENDPOINT"); ep != "" {
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(ep)))
	} else if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		exp, err = otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(ep), otlptracehttp.WithInsecure())
	}
	if err != nil || exp == nil {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(samplerFromEnv()),
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxQueueSize(4096),
			sdktrace.WithExportTimeout(5*time.Second),
		),
		sdktrace.WithResource(resourceFromEnv()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	return tp.Shutdown, nil
}
