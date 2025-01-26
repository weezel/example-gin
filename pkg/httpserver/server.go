package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"weezel/example-gin/pkg/ginmiddleware"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Option func(*HTTPServer)

func WithDefaultTracer(appName string) Option {
	return func(h *HTTPServer) {
		h.ginEngine.Use(otelgin.Middleware(appName))
		l.Logger.Info().Msg("Enabled tracing and metrics for gin gonic")
	}
}

func WithHTTPServer(httpServer *http.Server) Option {
	return func(h *HTTPServer) {
		h.httpServer = httpServer
	}
}

func WithHTTPAddr(host, port string) Option {
	return func(h *HTTPServer) {
		h.httpServer.Addr = net.JoinHostPort(host, port)
	}
}

func WithCustomHealthCheckHandler(healthCheckHandler gin.HandlerFunc) Option {
	return func(h *HTTPServer) {
		h.ginEngine.GET("/health", healthCheckHandler)
	}
}

type HTTPServer struct {
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

	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer cancel()

	r := gin.New()
	httpServer := &HTTPServer{
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
