package main

import (
	"context"
	"errors"
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

func main() {
	ctx := context.Background()
	var err error

	// Load config
	cfg := config.Config{}
	if err = env.Parse(&cfg); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	var dbConn *pgxpool.Pool

	r := httpserver.New()
	routes.AddRoutes(r)
	srv := httpserver.Config(r)
	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		dbConn, err = db.New(ctx, cfg)
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
