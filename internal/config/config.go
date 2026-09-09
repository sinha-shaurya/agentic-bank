package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// configFile is read from the process working directory at startup.
const configFile = "config.yaml"

// AppConfig is the top-level shape of config.yaml.
type AppConfig struct {
	AppConfigObject Config `yaml:"app"`
}

// Config is the fully resolved runtime configuration. It is built once at
// startup and injected downwards — nothing re-reads the file after Load
// returns.
type Config struct {
	Env          string `yaml:"env"`
	Port         int    `yaml:"port"`
	DatabaseURL  string `yaml:"database_url"`
	LogLevel     string `yaml:"log_level"`
	GenesisToken string `yaml:"genesis_token"`
}

func getConfigFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current working directory:", err)
	}
	return filepath.Join(cwd, configFile)
}

// Load reads config.yaml and fails fast on anything missing or malformed, so a
// misconfigured instance never starts and gets traffic routed to it.
func Load() (Config, error) {
	data, err := os.ReadFile(getConfigFilePath())
	if err != nil {
		return Config{}, fmt.Errorf("config: read %s: %w", configFile, err)
	}

	var file AppConfig
	if err := yaml.Unmarshal(data, &file); err != nil {
		return Config{}, fmt.Errorf("config: parse %s: %w", configFile, err)
	}

	cfg := file.AppConfigObject
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("config: app.database_url is required")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("config: app.port must be a valid port number, got %d", cfg.Port)
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("config: app.log_level must be one of debug|info|warn|error, got %q", cfg.LogLevel)
	}

	return nil
}

// Address is what the HTTP server binds to. It is deliberately 0.0.0.0 and not
// localhost: a container-local listener is unreachable from outside the
// container.
func (cfg Config) Address() string {
	return fmt.Sprintf("0.0.0.0:%d", cfg.Port)
}

func (cfg Config) IsProduction() bool {
	return cfg.Env == "production"
}
