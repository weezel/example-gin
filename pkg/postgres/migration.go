package postgres

import (
	"database/sql"
	"fmt"
	"weezel/example-gin/pkg/config"

	l "weezel/example-gin/pkg/logger"
)

func NewMigrationConnection(cfg config.Postgres) (*sql.DB, error) {
	psqlConfig := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.DBName,
		cfg.TLS)
	l.Logger.Debug().
		Str("username", cfg.Username).
		Str("password", cfg.Password[0:1]+"...").
		Str("hostname", cfg.Hostname).
		Str("port", cfg.Port).
		Str("dbname", cfg.DBName).
		Msg("Database migrate connection initialization")
	dbConn, err := sql.Open("pgx", psqlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return dbConn, err
}
