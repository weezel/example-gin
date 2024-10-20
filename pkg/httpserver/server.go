package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"weezel/example-gin/pkg/ginmiddleware"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func defaultTracer(ctx context.Context, serviceName, collectorAddress string) *sdktrace.TracerProvider {
	// Create the Jaeger exporter
	// exporter, err := otlptracehttp.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(collectorAddress))
	if err != nil {
		l.Logger.Fatal().Err(err).Msg("Failed to setup tracer")
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exporter),
		// Record information about this application in a Resource.
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(fmt.Sprintf("%s_%s", serviceName, l.UniqID())),
			),
		),
	)
	return tp
}

type Option func(*HTTPServer)

func WithTracer(tracer *sdktrace.TracerProvider, appName string) Option {
	return func(h *HTTPServer) {
		// Enable Opentelemetry tracing by default
		h.ginEngine.Use(otelgin.Middleware(appName))
		h.tracer = tracer
		otel.SetTracerProvider(h.tracer)
	}
}

// WithTracer allows to set up OTEL tracer.
// Example how to set it:
//
//	httpServer := httpserver.New(
//		httpserver.WithHTTPAddr(fmt.Sprintf("%s:%s", cfg.HTTPServer.Hostname, cfg.HTTPServer.Port)),
//		httpserver.WithDefaultTracer(ctx, "example-gin", os.Getenv("COLLECTOR_ADDR")),
//	)
func WithDefaultTracer(ctx context.Context, appName, collectorAddress string) Option {
	return func(h *HTTPServer) {
		h.ginEngine.Use(otelgin.Middleware(appName))
		h.tracer = defaultTracer(ctx, appName, collectorAddress)
		otel.SetTracerProvider(h.tracer)
	}
}

func WithHTTPServer(httpServer *http.Server) Option {
	return func(h *HTTPServer) {
		h.httpServer = httpServer
	}
}

func WithHTTPAddr(addr string) Option {
	return func(h *HTTPServer) {
		h.httpServer.Addr = addr
	}
}

func WithCustomHealthCheckHandler(healthCheckHandler gin.HandlerFunc) Option {
	return func(h *HTTPServer) {
		h.ginEngine.GET("/health", healthCheckHandler)
	}
}

type HTTPServer struct {
	tracer     *sdktrace.TracerProvider
	ginEngine  *gin.Engine
	httpServer *http.Server
}

// New returns a new HTTP server with custom configurations, like structured logging middleware.
// This is a general implementation that can be used in any server.
// Leverages options pattern.
func New(opts ...Option) *HTTPServer {
	if strings.ToLower(os.Getenv("DEBUG")) != "true" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	httpServer := &HTTPServer{
		tracer:    nil,
		ginEngine: r,
		httpServer: &http.Server{
			ReadTimeout: 60 * time.Second, // Mitigation against Slow loris attack (value from nginx)
			Addr:        net.JoinHostPort("0.0.0.0", "8080"),
			Handler:     r,
		},
	}

	// Use our own logging middleware
	r.Use(ginmiddleware.DefaultStructuredLogger())
	// Set secure headers
	r.Use(ginmiddleware.SecureHeaders())

	// Don't log health checks
	r.GET("/health", healthCheckHandler)

	r.Use(gin.Recovery())

	// Override defaults if any options are given
	for _, opt := range opts {
		opt(httpServer)
	}

	return httpServer
}

// Initializing the server in a goroutine so that it won't block
func (h *HTTPServer) Start() {
	l.Logger.Info().Msgf("Starting web server on %s", h.httpServer.Addr)
	go func() {
		if err := h.httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			l.Logger.Error().Err(err).Msg("HTTP server closed")
		}
	}()
}

func (h *HTTPServer) Shutdown(ctx context.Context) {
	timeout := 5 * time.Second
	cCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	l.Logger.Info().Msgf("Closing HTTP server with %s timeout", timeout)

	// There's no constant for context cancellation in tracer's context, therefore use dynamically created error
	if h.tracer != nil {
		if err := h.tracer.Shutdown(cCtx); err != nil &&
			errors.Is(err, errors.New("context canceled")) {
			l.Logger.Error().Err(err).Msg("Failed to shutdown opentelemetry tracer")
		}
	}

	if err := h.httpServer.Shutdown(cCtx); err != nil {
		l.Logger.Fatal().Err(err).Msg("Forced shutdown")
	}
}

func (h *HTTPServer) NewRouterGroup(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return h.ginEngine.Group(path, handlers...)
}

func (h *HTTPServer) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(httpMethod, relativePath, handlers...)
}

func (h *HTTPServer) POST(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodPost, relativePath, handlers...)
}

func (h *HTTPServer) GET(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodGet, relativePath, handlers...)
}

func (h *HTTPServer) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodDelete, relativePath, handlers...)
}

func (h *HTTPServer) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodPatch, relativePath, handlers...)
}

func (h *HTTPServer) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodPut, relativePath, handlers...)
}

func (h *HTTPServer) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodOptions, relativePath, handlers...)
}

func (h *HTTPServer) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Handle(http.MethodHead, relativePath, handlers...)
}

func (h *HTTPServer) Any(relativePath string, handlers ...gin.HandlerFunc) {
	h.ginEngine.Any(relativePath, handlers...)
}
