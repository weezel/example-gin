package profiling

import (
	"cmp"
	"context"
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // It's purposely exposed
	"os"
	"sync"
	"time"

	l "weezel/example-gin/pkg/logger"
)

type PprofServer struct {
	server     *http.Server
	once       *sync.Once
	listenAddr string
}

// NewPprofServer provides new debug http server
func NewPprofServer() *PprofServer {
	hostname := cmp.Or(os.Getenv("TRACE_SERVER_HOST"), "127.0.0.1")
	port := cmp.Or(os.Getenv("TRACE_SERVER_PORT"), "1337")
	listenAddress := net.JoinHostPort(hostname, port)
	return &PprofServer{
		listenAddr: listenAddress,
		server: &http.Server{
			Addr:              listenAddress,
			Handler:           http.DefaultServeMux,
			ReadHeaderTimeout: time.Second * 30,
		},
	}
}

func (p *PprofServer) Start() {
	go func() {
		l.Logger.Info().Msgf("Starting pprofiling server on %s", p.listenAddr)
		err := p.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Logger.Error().Err(err).Msg("Pprofiling server closed")
		}
	}()
}

func (p *PprofServer) Shutdown(ctx context.Context) {
	p.once.Do(func() {
		cCtx, cancel := context.WithTimeoutCause(
			ctx,
			time.Second*3,
			errors.New("debug server shutdown timeout"),
		)
		defer cancel()
		l.Logger.Info().Msg("Closing pprofiling server")
		if err := p.server.Shutdown(cCtx); err != nil {
			l.Logger.Error().Err(err).Msg("Failed to shutdown profiling server")
		}
	})
}
