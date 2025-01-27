package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"weezel/example-gin/cmd/webserver/routes"
	"weezel/example-gin/pkg/config"
	"weezel/example-gin/pkg/generated/sqlc"
	"weezel/example-gin/pkg/httpserver"
	"weezel/example-gin/pkg/postgres"
	"weezel/example-gin/pkg/tracer"

	l "weezel/example-gin/pkg/logger"
)

var serviceName = cmp.Or(os.Getenv("APP_NAME"), "example-gin")

// These will be filled in the build time by -X flag
var (
	Version   string
	BuildTime string
)

// Flags
var (
	showVersion     bool
	enableProfiling bool
)

func profiling() (*os.File, func()) {
	f, err := os.Create("cpu.prof")
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Couldn't create cpu.prof file")
	}

	if err = pprof.StartCPUProfile(f); err != nil {
		l.Logger.Panic().Err(err).Msg("Couldn't start CPU profile")
	}

	return f, pprof.StopCPUProfile
}

func main() {
	ctx := context.Background()
	var err error

	flag.BoolVar(&showVersion, "v", false, "Show version and build time")
	flag.BoolVar(&enableProfiling, "p", false, "Enable run time profiling")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s Build time: %s\n", Version, BuildTime)
		os.Exit(0)
	}

	l.Logger.Info().
		Str("version", Version).
		Str("build_time", BuildTime).
		Msg("Current build")

	if enableProfiling {
		profFHandle, profilingClose := profiling()
		defer profilingClose()
		defer func() {
			if err = profFHandle.Close(); err != nil {
				l.Logger.Error().Err(err).Msg("Failed to close cpu.prof file handle")
			}
		}()
	}

	// Load config
	cfg := config.Config{}
	if err = cfg.Parse(); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	otelTracerMetrics := tracer.NewOtelTracerMetrics(
		ctx,
		serviceName,
		"localhost",
		"4317",
		tracer.OtelTracingEnabled|tracer.OtelMetricsEnabled,
	)
	if err = otelTracerMetrics.Connect(ctx); err != nil {
		l.Logger.Panic().Err(err).Msg("OTEL client connection failed")
	}
	defer func() {
		otelTracerMetrics.Close(ctx)
	}()

	dbCtrl := postgres.New(
		cfg.Postgres,
		serviceName,
		postgres.WithTelemetryEnabled(),
	)
	err = dbCtrl.Connect(ctx)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Database connection failed")
	}
	defer dbCtrl.Close(ctx)
	queries := sqlc.New(dbCtrl.Pool())

	httpServer := httpserver.New(
		httpserver.WithHTTPAddr(cfg.HTTPServer.Hostname, cfg.HTTPServer.Port),
		httpserver.WithTelemetryEnabled(serviceName),
	)
	defer httpServer.Shutdown(ctx)

	routes.AddRoutes(httpServer, queries)

	// Starts webserver on Goroutine.
	httpServer.Start()

	// Set signal handler for handling the graceful shutdown
	sig := make(chan os.Signal, 1)
	// React to SIGINT and SIGTERM only
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	recvSignal := <-sig
	l.Logger.Info().Str("received_signal", recvSignal.String()).Msg("Performing a graceful shutdown")

	httpServer.Shutdown(ctx)

	l.Logger.Info().Msgf("Service %s exiting", serviceName)
}
