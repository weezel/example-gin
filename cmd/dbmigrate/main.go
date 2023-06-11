package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"weezel/example-gin/pkg/config"

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
	migrateAll  bool
	rollbackAll bool
	showStatus  bool
	showVersion bool
	wd          string //nolint:unused // This is an inner variable affect working directory
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

func createDatabase(ctx context.Context, cfg config.Postgres, dbName string) error {
	// Intentionally use hardcoded postgres username and database here
	// For RDS username is the specially created admin user
	psqlConfig := fmt.Sprintf("user=postgres password=%s host=%s port=%s dbname=postgres sslmode=disable",
		cfg.Password,
		cfg.Hostname,
		cfg.Port)
	l.Logger.Debug().
		Str("username", cfg.Username).
		Str("password", cfg.Password[0:1]+"...").
		Str("hostname", cfg.Hostname).
		Str("port", cfg.Port).
		Str("dbname", "postgres").
		Msg("Database creation connection")
	dbConn, err := sql.Open("pgx", psqlConfig)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to open database connection")
	}
	defer dbConn.Close()

	_, err = dbConn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s OWNER postgres ENCODING UTF8", dbName))
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	flag.BoolVar(&migrateAll, "m", false, "Run all migrations")
	flag.BoolVar(&rollbackAll, "r", false, "Rollback all migrations")
	flag.BoolVar(&showStatus, "s", false, "Show status of migrations")
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

	if !migrateAll && !rollbackAll && !showStatus {
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.Postgres{}
	if err := cfg.Parse(); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	psqlConfig := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.DBName)
	l.Logger.Debug().
		Str("username", cfg.Username).
		Str("password", cfg.Password[0:1]+"...").
		Str("hostname", cfg.Hostname).
		Str("port", cfg.Port).
		Str("dbname", cfg.DBName).
		Msg("Database configuration")
	dbConn, err := sql.Open("pgx", psqlConfig)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to open database connection")
	}
	defer dbConn.Close()

	if err = createDatabase(ctx, cfg, cfg.DBName); err != nil {
		l.Logger.Panic().Err(err).Msg("Error creating database")
	}

	if err = dbConn.PingContext(ctx); err != nil {
		l.Logger.Panic().Err(err).Msg("Ping database")
	}

	goose.SetBaseFS(sqlMigrations)

	if showStatus {
		if err = goose.Status(dbConn, schemasDir); err != nil {
			l.Logger.Error().Err(err).Msg("Database migration status")
		}
		return
	}

	if migrateAll && !rollbackAll {
		// Do the DB Migrations
		if err = goose.Status(dbConn, schemasDir); err != nil {
			l.Logger.Error().Err(err).Msg("Database migration status")
			return
		}
		if err = goose.Up(dbConn, schemasDir); err != nil {
			l.Logger.Error().Err(err).Msg("Migrations failed")
			return
		}
	} else if rollbackAll && !migrateAll {
		l.Logger.Info().Msg("Rollbacking the database migrations")
		// Rollback all the migrations until they are gone
		for {
			if err = goose.Down(dbConn, schemasDir); err != nil {
				l.Logger.Error().Err(err).Msg("Rollback failed")
				break
			}
		}
		l.Logger.Info().Msg("Rollbacks completed")
	}
	l.Logger.Info().Msg("Database migration completed")
}
