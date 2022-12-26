package config

type Config struct {
	DBHostname   string `env:"DB_HOST" envDefault:"localhost"`
	DBPort       string `env:"DB_PORT" envDefault:"5432"`
	DBName       string `env:"DB_NAME"`
	DBUsername   string `env:"DB_USERNAME"`
	DBPassword   string `env:"DB_PASSWORD,unset"` // XXX Use SecretsManager in AWS
	HTTPPort     string `env:"HTTP_PORT" envDefault:"8080"`
	HTTPHostname string `env:"HTTP_HOSTNAME" envDefault:"localhost"`
}
