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
	"weezel/example-gin/pkg/db"
	"weezel/example-gin/pkg/httpserver"

	l "weezel/example-gin/pkg/logger"

	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v4/pgxpool"
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
	if err = env.Parse(&cfg); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	var dbConn *pgxpool.Pool
	// This must be before httpserver.New() since that sets the database pointer to
	// Gin context.
	dbConn, err = db.New(ctx, cfg)
	if err != nil {
		l.Logger.Fatal().Err(err).Msg("Database connection failed")
	}

	engine := httpserver.New()
	routes.AddRoutes(engine)
	srv := httpserver.Config(engine)
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
