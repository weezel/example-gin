package main

import (
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

	l "weezel/example-gin/pkg/logger"
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

	f, err := os.Create("cpu.prof")
	if err != nil {
		l.Logger.Fatal().Err(err).Msg("Couldn't create cpu.prof file")
	}
	defer func() {
		if err = f.Close(); err != nil {
			l.Logger.Error().Err(err).Msg("Failed to close cpu.prof file handle")
		}
	}()
	if err = pprof.StartCPUProfile(f); err != nil {
		l.Logger.Fatal().Err(err).Msg("Couldn't start CPU profile")
	}
	defer pprof.StopCPUProfile()

	// Load config
	cfg := config.Config{}
	if err = cfg.Parse(); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	dbCtrl := postgres.New(
		postgres.WithUsername(cfg.Postgres.Username),
		postgres.WithPassword(cfg.Postgres.Password),
		postgres.WithPort(cfg.Postgres.Port),
		postgres.WithDBName(cfg.Postgres.DBName),
		postgres.WithSSLMode(postgres.SSLModeDisable), // This is running on localhost only
		postgres.WithApplicationName("example-gin"),
	)
	err = dbCtrl.Connect(ctx)
	if err != nil {
		l.Logger.Fatal().Err(err).Msg("Database connection failed")
	}
	defer dbCtrl.Close(ctx)
	queries := sqlc.New(dbCtrl.Pool())

	httpServer := httpserver.New(
		httpserver.WithHTTPAddr(cfg.HTTPServer.Hostname, cfg.HTTPServer.Port),
		// httpserver.WithDefaultTracer(ctx, "example-gin", os.Getenv("COLLECTOR_ADDR")),
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

	l.Logger.Info().Msgf("Service %s exiting", os.Getenv("APP_NAME"))
}
