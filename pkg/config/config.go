package config

import (
	"fmt"

	"github.com/caarlos0/env/v8"
)

// TODO Separation
type Postgres struct {
	Hostname string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	DBName   string `env:"DB_NAME"`
	Username string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD,unset"` // XXX Use SecretsManager in AWS
	TLS      string `env:"DB_TLS" envDefault:"prefer"`
}

func (p *Postgres) Parse() error {
	return env.Parse(p)
}

type HTTPServer struct {
	Port     string `env:"HTTP_PORT" envDefault:"8080"`
	Hostname string `env:"HTTP_HOSTNAME" envDefault:"localhost"`
}

func (h *HTTPServer) Parse() error {
	return env.Parse(h)
}

type Config struct {
	HTTPServer HTTPServer
	Postgres   Postgres
}

func (c *Config) Parse() error {
	var err error

	if err = c.HTTPServer.Parse(); err != nil {
		return fmt.Errorf("failed to parse HTTP server configs: %w", err)
	}
	if err = c.Postgres.Parse(); err != nil {
		return fmt.Errorf("failed to parse Postgres configs: %w", err)
	}

	return nil
}
