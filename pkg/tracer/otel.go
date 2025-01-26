package tracer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	l "weezel/example-gin/pkg/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OtelMode defines which OTEL modes will be enable: metrics, tracing or logging.
type OtelMode uint16

const (
	OtelMetricsEnabled OtelMode = 1 << iota
	OtelTracingEnabled
	OtelLoggingEnabled
)

type OtelTracerMetrics struct {
	metrics     *sdkmetric.MeterProvider
	tracer      *sdktrace.TracerProvider
	res         *resource.Resource
	connection  *grpc.ClientConn
	closeOnce   *sync.Once
	serviceName attribute.KeyValue
	modes       OtelMode
}

func NewOtelTracerMetrics(
	ctx context.Context,
	appName string,
	collectorHost string,
	collectorPort string,
	modes OtelMode,
) *OtelTracerMetrics {
	serviceName := semconv.ServiceNameKey.String(appName)
	// serviceName := semconv.ServiceNameKey.String(fmt.Sprintf("%s_%s", appName, l.UniqID()))
	res, err := resource.New(ctx, resource.WithAttributes(serviceName))
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Couldn't create a new tracing resource")
	}

	otelTracerMetrics := &OtelTracerMetrics{
		serviceName: serviceName,
		res:         res,
		modes:       modes,
	}

	otelTracerMetrics.connection, err = grpc.NewClient(
		net.JoinHostPort(collectorHost, collectorPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to create gRPC connection to collector")
	}

	return otelTracerMetrics
}

func (o *OtelTracerMetrics) Connect(ctx context.Context) error {
	for bit := OtelMetricsEnabled; bit <= OtelLoggingEnabled; bit <<= 1 {
		if o.modes&bit != 0 {
			switch bit {
			case OtelMetricsEnabled:
				if err := o.defaultMetricsProvider(ctx); err != nil {
					return fmt.Errorf("initialize metrics: %w", err)
				}
			case OtelTracingEnabled:
				if err := o.defaultTracerProvider(ctx); err != nil {
					return fmt.Errorf("initialize tracer: %w", err)
				}
			case OtelLoggingEnabled:
				l.Logger.Panic().Msg("Logging is not implemented yet")
			}
		}
	}

	return nil
}

func (o *OtelTracerMetrics) Close(ctx context.Context) {
	o.closeOnce.Do(func() {
		var connErr error
		var tracerErr error
		var metricsErr error
		if o.connection != nil {
			connErr = o.connection.Close()
		}
		if o.tracer != nil {
			tracerErr = o.tracer.Shutdown(ctx)
		}
		if o.metrics != nil {
			metricsErr = o.metrics.Shutdown(ctx)
		}
		errs := errors.Join(connErr, tracerErr, metricsErr)
		if errs != nil {
			l.Logger.Error().Err(errs).Msg("Error closing otel tracer metrics")
		}
	})
}

func (o *OtelTracerMetrics) defaultTracerProvider(ctx context.Context) error {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(o.connection))
	if err != nil {
		return fmt.Errorf("trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	o.tracer = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(o.res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(o.tracer)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	l.Logger.Info().Msg("Enabled global tracing")

	return nil
}

func (o *OtelTracerMetrics) defaultMetricsProvider(ctx context.Context) error {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(o.connection))
	if err != nil {
		return fmt.Errorf("metrics exporter: %w", err)
	}

	o.metrics = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(o.res),
	)
	otel.SetMeterProvider(o.metrics)

	l.Logger.Info().Msg("Enabled global metrics")

	return nil
}
