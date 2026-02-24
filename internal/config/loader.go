package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func Load() *Config {
	loadEnvFile()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse environment variables: %v", err)
	}

	if err := validate(cfg); err != nil {
		log.Fatalf("configuration validation failed: %v", err)
	}

	return cfg
}

func loadEnvFile() {
	if os.Getenv("APP_ENV") == "production" {
		return
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("warning: could not determine current file path, skipping .env load")
		return
	}

	baseDir := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	envPath := filepath.Join(baseDir, ".env")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Printf("warning: .env file not found at %s, using system environment variables", envPath)
		return
	}

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("warning: failed to load .env file: %v", err)
		return
	}

	log.Printf("loaded environment from %s", envPath)
}

func LoadOrDefault() *Config {
	cfg := &Config{
		ServerPort:        "8080",
		GinMode:           "test",
		AppEnv:            "test",
		DBHost:            "localhost",
		DBPort:            "5432",
		DBUser:            "test_user",
		DBPassword:        "test_password",
		DBName:            "test_db",
		RedisHost:         "localhost",
		RedisPort:         "6379",
		RedisPassword:     "",
		RedisDB:           0,
		ShutdownTimeout:   30 * time.Second,
		URLExpirationDays: 30,
		LogLevel:          "debug",
		LogFormat:         "console",
	}
	return cfg
}
