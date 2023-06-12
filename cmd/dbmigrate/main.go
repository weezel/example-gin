package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"weezel/example-gin/pkg/config"
	"weezel/example-gin/pkg/postgres"

	l "weezel/example-gin/pkg/logger"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

// These will be filled in the build time by -X flag
var (
	Version   string
	BuildTime string
)

// Flags
var (
	showVersion bool
	wd          string //nolint:unused // This is an inner variable affects working directory
)

var schemasDir = "schemas"

//go:embed schemas/*.sql
var sqlMigrations embed.FS

func init() {
	log.SetFlags(0)

	var err error
	wd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

func createDatabase(ctx context.Context, cfg config.Postgres) error {
	// Intentionally use hardcoded postgres username and database here.
	// For RDS username is the specially created admin user
	newCfg := cfg
	newCfg.Username = "postgres"
	newCfg.DBName = "postgres"
	dbConn, err := postgres.NewMigrationConnection(newCfg)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to open database connection")
	}
	defer dbConn.Close()

	_, err = dbConn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s OWNER postgres ENCODING UTF8", cfg.DBName))
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

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

	cfg := config.Postgres{}
	if err := cfg.Parse(); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	if err := createDatabase(ctx, cfg); err != nil {
		l.Logger.Panic().Err(err).Msg("Error creating database")
	}

	dbConn, err := postgres.NewMigrationConnection(cfg)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to initialize database connection for migrations")
	}
	defer dbConn.Close()

	if err = dbConn.PingContext(ctx); err != nil {
		l.Logger.Panic().Err(err).Msg("Ping database")
	}

	goose.SetBaseFS(sqlMigrations)

	if os.Getenv("DO_ROLLBACK") != "" {
		// Do the DB Migrations
		if err = goose.Status(dbConn, schemasDir); err != nil {
			l.Logger.Error().Err(err).Msg("Database migration status")
			return
		}
		l.Logger.Info().Msg("Rollbacking the database migrations")
		// Rollback all the migrations until they are gone
		for {
			if err = goose.Down(dbConn, schemasDir); err != nil {
				l.Logger.Error().Err(err).Msg("Rollback failed")
				break
			}
		}
		l.Logger.Info().Msg("Rollbacks completed")
	} else {
		// Do the DB Migrations
		if err = goose.Status(dbConn, schemasDir); err != nil {
			l.Logger.Error().Err(err).Msg("Database migration status")
			return
		}
		if err = goose.Up(dbConn, schemasDir); err != nil {
			l.Logger.Error().Err(err).Msg("Migrations failed")
			return
		}
	}

	l.Logger.Info().Msg("Database migration completed")
}
