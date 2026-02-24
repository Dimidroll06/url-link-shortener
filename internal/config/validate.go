package config

import (
	"fmt"
	"strings"
)

func validate(cfg *Config) error {
	var errs []string

	if cfg.IsProduction() && strings.TrimSpace(cfg.DBPassword) == "" {
		errs = append(errs, "DB_PASSWORD is required in production environment")
	}

	if cfg.ServerPort == "" {
		errs = append(errs, "SERVER_PORT cannot be empty")
	}

	validGinModes := map[string]bool{
		"debug":   true,
		"release": true,
		"test":    true,
	}
	if !validGinModes[cfg.GinMode] {
		errs = append(errs, fmt.Sprintf("GIN_MODE must be one of: debug, release, test (got: %s)", cfg.GinMode))
	}

	validLogFormats := map[string]bool{
		"json":    true,
		"console": true,
	}
	if !validLogFormats[cfg.LogFormat] {
		errs = append(errs, fmt.Sprintf("LOG_FORMAT must be one of: json, console (got: %s)", cfg.LogFormat))
	}

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"panic": true,
		"fatal": true,
	}

	if !validLevels[cfg.LogLevel] {
		errs = append(errs, fmt.Sprintf("LOG_LEVEL must be one of: debug, info, warn, error, panic, fatal (got: %s)", cfg.LogLevel))
	}

	if cfg.URLExpirationDays <= 0 {
		errs = append(errs, "URL_EXPIRATION_DAYS must be positive")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration errors:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
