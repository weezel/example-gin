package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"weezel/example-gin/pkg/config"

	l "weezel/example-gin/pkg/logger"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDatabaseRetriesExceeded = errors.New("database retries exceeded")

type SSLModes string

var (
	SSLModeDisable    SSLModes = "disable"
	SSLModeAllow      SSLModes = "allow"
	SSLModePrefer     SSLModes = "prefer"
	SSLModeRequire    SSLModes = "require"
	SSLModeVerifyCA   SSLModes = "verify-ca"
	SSLModeVerifyFull SSLModes = "verify-full"
)

type Contollerer interface {
	Connect(context.Context) error
}

type Controller struct {
	pool           *pgxpool.Pool
	config         *pgxpool.Config
	maxConnRetries uint
}

type Option func(*Controller)

func New(cfg config.Postgres, appName string, opts ...Option) *Controller {
	switch cfg.TLS {
	case string(SSLModeDisable),
		string(SSLModeAllow),
		string(SSLModePrefer),
		string(SSLModeRequire),
		string(SSLModeVerifyCA),
		string(SSLModeVerifyFull):
		break
	default:
		err := fmt.Errorf("invalid SSLMode: %s", cfg.TLS)
		l.Logger.Panic().Err(err).Msg("Unsupported SSL mode given")
	}

	//nolint:nosprintfhostport // False positive, this isn't a regular hostname:port syntax here
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&application_name=%s",
		cfg.Username,
		url.QueryEscape(cfg.Password),
		cfg.Hostname,
		cfg.Port,
		cfg.DBName,
		cfg.TLS,
		appName,
	)
	dbConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		l.Logger.Panic().Err(err).Msg("Cannot parse database config")
	}

	ctrl := &Controller{config: dbConfig}
	ctrl.config.MaxConnLifetime = 10 * time.Minute
	ctrl.config.MaxConnIdleTime = 10 * time.Second
	ctrl.config.HealthCheckPeriod = 30 * time.Second
	ctrl.config.MaxConns = 5
	ctrl.config.MinConns = 1

	// Override defaults
	for _, opt := range opts {
		opt(ctrl)
	}

	// If there's a "proxy" in hostname, use QueryExecMode to avoid surprises
	if strings.Contains(strings.ToLower(cfg.Hostname), "proxy") {
		ctrl.config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	return ctrl
}

func WithTelemetryEnabled() Option {
	return func(pc *Controller) {
		pc.config.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithIncludeQueryParameters())
	}
}

func WithPoolMaxConns(poolMaxConns int32) Option {
	return func(pc *Controller) {
		pc.config.MaxConns = poolMaxConns
	}
}

func (c *Controller) Close(ctx context.Context) {
	timeout := 5 * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	l.Logger.Info().Msgf("Closing database connections with %s timeout", timeout)

	// If there are still acquired connections, close them.
	// Otherwise pool.Close() will wait until the timeouts are reached.
	idleConns := c.pool.AcquireAllIdle(timeoutCtx)
	for _, conn := range idleConns {
		conn.Release()
	}
	c.pool.Close()
}

func (c *Controller) Pool() *pgxpool.Pool {
	return c.pool
}

func (c *Controller) Connect(ctx context.Context) error {
	started := time.Now()
	var retries uint
	var err error
	for {
		c.pool, err = pgxpool.NewWithConfig(ctx, c.config)
		if err != nil {
			l.Logger.Error().Err(err).Msg("Couldn't connect to database")
		}

		if err = c.pool.Ping(ctx); err == nil {
			break
		}

		delay := math.Ceil(math.Pow(2, float64(retries)))
		time.Sleep(time.Duration(delay) * time.Second)
		retries++

		l.Logger.Warn().Msgf("Retrying db connection %d/%d (%s since started)",
			retries, c.maxConnRetries, time.Since(started))

		if retries > c.maxConnRetries {
			return fmt.Errorf("%w [%d/%d]",
				ErrDatabaseRetriesExceeded,
				retries,
				c.maxConnRetries,
			)
		}
	}

	return nil
}

// NewMigrationConnection opens a new connection for database migrations
func NewMigrationConnection(cfg config.Postgres) (*sql.DB, error) {
	//nolint:nosprintfhostport // False positive, cannot use net.JoinHostPort() here
	psqlConfig := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&application_name=%s",
		cfg.Username,
		url.QueryEscape(cfg.Password),
		cfg.Hostname,
		cfg.Port,
		cfg.DBName,
		cfg.TLS,
		"example-gin migrations",
	)
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

	return dbConn, nil
}
