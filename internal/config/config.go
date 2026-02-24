package config

import (
	"fmt"
	"time"
)

type Config struct {
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`
	GinMode    string `env:"GIN_MODE" envDefault:"release"`
	AppEnv     string `env:"APP_ENV" envDefault:"development"`

	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:""`
	DBName     string `env:"DB_NAME" envDefault:"postgres"`

	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
	URLExpirationDays int           `env:"URL_EXPIRATION_DAYS" envDefault:"30"`

	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func (c *Config) RedisURL() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}
