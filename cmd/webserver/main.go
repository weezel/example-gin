package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weezel/example-gin/cmd/webserver/routes"
	"weezel/example-gin/pkg/config"
	"weezel/example-gin/pkg/generated/sqlc"
	"weezel/example-gin/pkg/httpserver"
	"weezel/example-gin/pkg/postgres"

	l "weezel/example-gin/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

// These will be filled in the build time by -X flag
var (
	Version   string
	BuildTime string
)

// Flags
var (
	showVersion bool
)

func main() {
	ctx := context.Background()
	var err error

	flag.BoolVar(&showVersion, "v", false, "Show version and build time")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s Build time: %s\n", Version, BuildTime)
		os.Exit(0)
	}

	l.Logger.Info().
		Str("version", Version).
		Str("build_time", BuildTime).
		Msg("Current build")

	// Load config
	cfg := config.Config{}
	if err = cfg.Parse(); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	var dbConn *pgxpool.Pool
	// This must be before httpserver.New() since that sets the database pointer to
	// Gin context.
	dbConn, err = postgres.New(ctx, cfg.Postgres)
	if err != nil {
		l.Logger.Fatal().Err(err).Msg("Database connection failed")
	}
	queries := sqlc.New(dbConn)

	// Create a new HTTP engine and Opentelemetry Trace provider.
	// Trace provider uses Gin's middleware so it's omnipresent.
	engine, traceProvider := httpserver.New()
	defer func() {
		// There's no constant for context cancellation in
		// tracer's context, therefore use dynamicly created error
		if err = traceProvider.Shutdown(ctx); err != nil &&
			errors.Is(err, errors.New("context canceled")) {
			l.Logger.Error().Err(err).Msg("Failed to shutdown opentelemetry tracer")
		}
	}()
	otel.SetTracerProvider(traceProvider)

	routes.AddRoutes(engine, queries)
	srv := httpserver.Config(engine, cfg.HTTPServer)
	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err = srv.ListenAndServe(); err != nil &&
			errors.Is(err, http.ErrServerClosed) {
			l.Logger.Error().Err(err).Msg("HTTP server closed")
		}
	}()

	if err != nil {
		l.Logger.Error().Err(err).Msg("Failed to connect to database")
	}

	sig := make(chan os.Signal, 1)
	// React to SIGINT and SIGTERM only
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	l.Logger.Info().Msg("Shutting down the service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown connections to database
	dbConn.Close()

	if err := srv.Shutdown(ctx); err != nil {
		l.Logger.Fatal().Err(err).Msg("Forced shutdown")
	}

	l.Logger.Info().Msg("Service exiting")
}
