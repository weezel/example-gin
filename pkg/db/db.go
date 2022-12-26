package db

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
	"weezel/example-gin/pkg/config"

	l "weezel/example-gin/pkg/logger"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dbConnRetries  = 3
	postgresConfig = "sslmode=disable&pool_max_conns=5"
)

var (
	once   sync.Once
	dbPool *pgxpool.Pool
	dbErr  error
)

// New initializes database connection once. Also known as singleton.
func New(ctx context.Context, dbConf config.Config) (*pgxpool.Pool, error) {
	pgConfigURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?"+postgresConfig,
		dbConf.DBUsername,
		dbConf.DBPassword,
		dbConf.DBHostname,
		dbConf.DBPort,
		dbConf.DBName)
	var retries int
	once.Do(func() {
		dbPool, dbErr = pgxpool.Connect(ctx, pgConfigURL)
		if dbErr != nil {
			l.Logger.Fatal().Err(dbErr).Msg("Failed to start database")
		}

		started := time.Now()
		for {
			if dbErr = dbPool.Ping(ctx); dbErr == nil {
				break
			}

			dbPool, dbErr = pgxpool.Connect(ctx, pgConfigURL)
			if dbErr == nil {
				break
			}

			delay := math.Ceil(math.Pow(2, float64(retries)))
			time.Sleep(time.Duration(delay) * time.Second)
			retries++

			l.Logger.Warn().Msgf("Retrying db connection %d/%d (%s since started)",
				retries, dbConnRetries, time.Since(started))

		}
	})
	if retries > dbConnRetries {
		return nil, fmt.Errorf("Couldn't connect to database after %d retries", retries-1)
	}

	return dbPool, dbErr
}

func GetPool() *pgxpool.Pool {
	return dbPool
}
