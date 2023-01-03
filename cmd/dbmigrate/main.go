package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"weezel/example-gin/pkg/config"

	l "weezel/example-gin/pkg/logger"

	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	migrateAll  bool
	rollbackAll bool
	showStatus  bool
	wd          string
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

func main() {
	ctx := context.Background()

	flag.BoolVar(&migrateAll, "m", false, "Run all migrations")
	flag.BoolVar(&rollbackAll, "r", false, "Rollback all migrations")
	flag.BoolVar(&showStatus, "s", false, "Show status of migrations")
	flag.Parse()

	if !migrateAll && !rollbackAll && !showStatus {
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to parse config")
	}

	psqlConfig := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHostname,
		cfg.DBPort,
		cfg.DBName)
	l.Logger.Debug().
		Str("username", cfg.DBUsername).
		Str("password", cfg.DBPassword[0:3]+"...").
		Str("hostname", cfg.DBHostname).
		Str("port", cfg.DBPort).
		Str("dbname", cfg.DBName).
		Msg("Database configuration")
	dbConn, err := sql.Open("pgx", psqlConfig)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Failed to open database connection")
	}
	defer dbConn.Close()

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
