package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// configFile is read from the process working directory at startup.
const configFile = "config.yaml"

// defaultSSLMode is what an unset postgres.sslmode falls back to.
const defaultSSLMode = "disable"

// AppConfig is the top-level shape of config.yaml.
type AppConfig struct {
	AppConfigObject Config `yaml:"app"`
}

// Config is the fully resolved runtime configuration. It is built once at
// startup and injected downwards — nothing re-reads the file after Load
// returns.
type Config struct {
	Env          string   `yaml:"env"`
	Port         int      `yaml:"port"`
	Postgres     Postgres `yaml:"postgres"`
	LogLevel     string   `yaml:"log_level"`
	GenesisToken string   `yaml:"genesis_token"`
}

// Postgres describes the database connection. Host carries the port too, e.g.
// "localhost:5432".
type Postgres struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Debug    bool   `yaml:"debug"`
	SSLMode  string `yaml:"sslmode"`
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
	if cfg.Postgres.SSLMode == "" {
		cfg.Postgres.SSLMode = defaultSSLMode
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if err := cfg.Postgres.validate(); err != nil {
		return err
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

func (pg Postgres) validate() error {
	if pg.Host == "" {
		return fmt.Errorf("config: app.postgres.host is required")
	}

	// Host must be addressable as-is by the driver, so reject a missing port here
	// rather than at first connect.
	if _, _, err := net.SplitHostPort(pg.Host); err != nil {
		return fmt.Errorf("config: app.postgres.host must be host:port, got %q", pg.Host)
	}

	if pg.Username == "" {
		return fmt.Errorf("config: app.postgres.username is required")
	}

	if pg.DBName == "" {
		return fmt.Errorf("config: app.postgres.dbname is required")
	}

	switch pg.SSLMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("config: app.postgres.sslmode is invalid, got %q", pg.SSLMode)
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
