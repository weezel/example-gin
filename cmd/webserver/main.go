package main

import (
	"context"
	"flag"
	"fmt"
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
	dbConn, err = postgres.New(ctx, cfg.Postgres)
	if err != nil {
		l.Logger.Fatal().Err(err).Msg("Database connection failed")
	}
	defer dbConn.Close()
	queries := sqlc.New(dbConn)

	httpServer := httpserver.New(
		httpserver.WithHTTPAddr(fmt.Sprintf("%s:%s", cfg.HTTPServer.Hostname, cfg.HTTPServer.Port)),
		httpserver.WithDefaultTracer(ctx, "example-gin", os.Getenv("COLLECTOR_ADDR")),
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	httpServer.Shutdown(ctx)

	l.Logger.Info().Msgf("Service %s exiting", os.Getenv("APP_NAME"))
}
