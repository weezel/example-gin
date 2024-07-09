package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"time"

	"weezel/example-gin/pkg/config"

	l "weezel/example-gin/pkg/logger"

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
	pool            *pgxpool.Pool
	username        string
	password        string
	hostname        string
	port            string
	dbName          string
	applicationName string
	dbURL           string
	sslMode         SSLModes
	poolMaxConns    uint
	maxConnRetries  uint
}

type Option func(*Controller)

func New(opts ...Option) *Controller {
	ctrl := &Controller{
		port:           "5432",
		maxConnRetries: 5,
		poolMaxConns:   5,
		username:       "postgres",
		hostname:       "localhost",
		dbName:         "postgres",
		sslMode:        SSLModePrefer,
	}

	for _, opt := range opts {
		opt(ctrl)
	}

	//nolint:nosprintfhostport // False positive, cannot use net.JoinHostPort() here
	ctrl.dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d&application_name=%s",
		ctrl.username,
		url.QueryEscape(ctrl.password),
		ctrl.hostname,
		ctrl.port,
		ctrl.dbName,
		ctrl.sslMode,
		ctrl.poolMaxConns,
		ctrl.applicationName,
	)

	return ctrl
}

func WithUsername(username string) Option {
	return func(pc *Controller) {
		pc.username = username
	}
}

func WithPassword(password string) Option {
	return func(pc *Controller) {
		pc.password = password
	}
}

func WithHostname(hostname string) Option {
	return func(pc *Controller) {
		pc.hostname = hostname
	}
}

func WithPort(port string) Option {
	return func(pc *Controller) {
		pc.port = port
	}
}

func WithDBName(dbName string) Option {
	return func(pc *Controller) {
		pc.dbName = dbName
	}
}

func WithSSLMode(sslMode SSLModes) Option {
	switch sslMode {
	case SSLModeDisable, SSLModeAllow, SSLModePrefer, SSLModeRequire, SSLModeVerifyCA, SSLModeVerifyFull:
		break
	default:
		err := fmt.Errorf("invalid SSLMode: %s", sslMode)
		l.Logger.Fatal().Err(err).Msg("Unsupported SSL mode given")
	}

	return func(pc *Controller) {
		pc.sslMode = sslMode
	}
}

func WithPoolMaxConns(poolMaxConns uint) Option {
	return func(pc *Controller) {
		pc.poolMaxConns = poolMaxConns
	}
}

func WithApplicationName(applicationName string) Option {
	return func(pc *Controller) {
		pc.applicationName = applicationName
	}
}

func (c *Controller) Close(ctx context.Context) {
	timeout := 5 * time.Second
	_, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	l.Logger.Info().Msgf("Closing database connections with %s timeout", timeout)

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
		c.pool, err = pgxpool.New(ctx, c.dbURL)
		if err != nil {
			l.Logger.Error().Err(err).Msg("Couldn't connect to DB")
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
