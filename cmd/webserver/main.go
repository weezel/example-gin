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
	"weezel/example-gin/pkg/httpserver"

	l "weezel/example-gin/pkg/logger"
)

func main() {
	r := httpserver.New()
	routes.AddRoutes(r)
	srv := httpserver.Config(r)
	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			errors.Is(err, http.ErrServerClosed) {
			l.Logger.Error().Err(err).Msg("HTTP server closed")
		}
	}()

	// TODO Graceful shutdown of database

	sig := make(chan os.Signal, 1)
	// React to SIGINT and SIGTERM only
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	l.Logger.Info().Msg("Shutting down the service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Logger.Fatal().Err(err).Msg("Forced shutdown")
	}

	l.Logger.Info().Msg("Service exiting")
}
