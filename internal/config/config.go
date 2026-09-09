package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the fully resolved runtime configuration. It is built once at
// startup from the environment and injected downwards — nothing reads os.Getenv
// after Load returns, and nothing reads a file relative to the working
// directory (which is not stable inside a container or on Render).
type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	LogLevel        string
	GenesisToken    string
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment and fails fast on anything
// missing or malformed, so a misconfigured instance never starts and gets
// traffic routed to it.
func Load() (Config, error) {
	cfg := Config{
		Env:          getEnv("APP_ENV", "development"),
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		LogLevel:     strings.ToLower(getEnv("LOG_LEVEL", "info")),
		GenesisToken: os.Getenv("GENESIS_TOKEN"),
	}

	shutdownTimeout, err := getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout = shutdownTimeout

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("config: PORT must be a valid port number, got %q", cfg.Port)
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("config: LOG_LEVEL must be one of debug|info|warn|error, got %q", cfg.LogLevel)
	}

	return nil
}

// Address is what the HTTP server binds to. It is deliberately 0.0.0.0 and not
// localhost: a container-local listener is unreachable from outside the
// container.
func (cfg Config) Address() string {
	return "0.0.0.0:" + cfg.Port
}

func (cfg Config) IsProduction() bool {
	return cfg.Env == "production"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a valid duration, got %q", key, value)
	}

	return duration, nil
}
