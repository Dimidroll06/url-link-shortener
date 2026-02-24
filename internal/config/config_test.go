package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadOrDefault(t *testing.T) {
	cfg := LoadOrDefault()

	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "test", cfg.GinMode)
	assert.Equal(t, "test", cfg.AppEnv)
	assert.Equal(t, 30, cfg.URLExpirationDays)
	assert.Equal(t, 30*time.Second, cfg.ShutdownTimeout)
}

func TestDatabaseURL(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "test_user",
		DBPassword: "test_pass",
		DBName:     "test_db",
	}

	expected := "postgres://test_user:test_pass@localhost:5432/test_db?sslmode=disable"
	assert.Equal(t, expected, cfg.DatabaseURL())
}

func TestRedisURL(t *testing.T) {
	cfg := &Config{
		RedisHost: "localhost",
		RedisPort: "6379",
	}

	expected := "localhost:6379"
	assert.Equal(t, expected, cfg.RedisURL())
}

func TestIsProduction(t *testing.T) {
	t.Run("production", func(t *testing.T) {
		cfg := &Config{AppEnv: "production"}
		assert.True(t, cfg.IsProduction())
		assert.False(t, cfg.IsDevelopment())
	})

	t.Run("development", func(t *testing.T) {
		cfg := &Config{AppEnv: "development"}
		assert.False(t, cfg.IsProduction())
		assert.True(t, cfg.IsDevelopment())
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid production config", func(t *testing.T) {
		cfg := &Config{
			AppEnv:            "production",
			DBPassword:        "password",
			ServerPort:        "8080",
			GinMode:           "release",
			LogFormat:         "json",
			LogLevel:          "info",
			URLExpirationDays: 30,
		}
		assert.NoError(t, validate(cfg))
	})

	t.Run("invalid GinMode", func(t *testing.T) {
		cfg := &Config{
			AppEnv:            "development",
			DBPassword:        "password",
			ServerPort:        "8080",
			GinMode:           "invalid",
			LogFormat:         "json",
			LogLevel:          "info",
			URLExpirationDays: 30,
		}
		assert.Error(t, validate(cfg))
	})

	t.Run("invalid LogLevel", func(t *testing.T) {
		cfg := &Config{
			AppEnv:            "production",
			DBPassword:        "password",
			ServerPort:        "8080",
			GinMode:           "release",
			LogFormat:         "json",
			LogLevel:          "invalid",
			URLExpirationDays: 30,
		}
		assert.Error(t, validate(cfg))
	})

	t.Run("invalid URLExpirationDays", func(t *testing.T) {
		cfg := &Config{
			AppEnv:            "production",
			DBPassword:        "password",
			ServerPort:        "8080",
			GinMode:           "release",
			LogFormat:         "json",
			LogLevel:          "info",
			URLExpirationDays: -1,
		}
		assert.Error(t, validate(cfg))
	})
}
