package httpserver

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"weezel/example-gin/pkg/config"
	"weezel/example-gin/pkg/ginmiddleware"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	serviceName = "example-gin"
	appEnv      = "production"
	jaegerURL   = "http://localhost:14268/api/traces"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create the Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exporter),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(fmt.Sprintf("%s_%s", serviceName, l.UniqID())),
			attribute.String("environment", appEnv),
		)),
	)
	return tp, nil
}

// New returns a new Gin engine with our custom configurations, like logging middleware.
// This is a general implementation that can be used in any server.
// Opentelemetry tracer is being hooked up as a middleware to get traces of the requests.
// Returns gin Engine and OpenTelemetry tracer.
func New() (*gin.Engine, *sdktrace.TracerProvider) {
	tp, err := initTracer()
	if err != nil {
		l.Logger.Error().Err(err).Msg("Failed to initialize opentelemetry tracer")
	}

	if strings.ToLower(os.Getenv("DEBUG")) != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// Enable Opentelemetry tracing by default
	r.Use(otelgin.Middleware("example-gin"))
	// Use our own logging middleware
	r.Use(ginmiddleware.DefaultStructuredLogger())
	// Database connection must be established at this point!
	r.Use(ginmiddleware.Postgres())
	r.Use(ginmiddleware.SecureHeaders())
	r.Use(gin.Recovery())

	return r, tp
}

func Config(r http.Handler, cfg config.HTTPServer) *http.Server {
	return &http.Server{
		ReadTimeout: 60 * time.Second, // Mitigation against Slow loris attack (value from nginx)
		Addr:        fmt.Sprintf("%s:%s", cfg.Hostname, cfg.Port),
		Handler:     r,
	}
}
