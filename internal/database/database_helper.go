// utility package for database setup
package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/sinha-shaurya/agentic-bank/internal/config"
)

const (
	maxIdleConns    = 5
	maxOpenConns    = 25
	connMaxLifetime = 5 * time.Minute
	slowThreshold   = time.Second
)

// DB owns the connection pool. It is built once at startup from injected
// config and handed to every controller.
type DB struct {
	gorm *gorm.DB
}

// New opens the pool and applies its connection properties.
func New(appConfig config.Config, logger *zap.Logger) (*DB, error) {
	gormDB, err := openConnection(appConfig)
	if err != nil {
		return nil, err
	}

	postgresDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("database: get sql handle: %w", err)
	}

	// set connection properties
	postgresDB.SetMaxIdleConns(maxIdleConns)
	postgresDB.SetMaxOpenConns(maxOpenConns)
	postgresDB.SetConnMaxLifetime(connMaxLifetime)

	logger.Info("db context loaded successfully",
		zap.String("host", appConfig.Postgres.Host),
		zap.String("dbname", appConfig.Postgres.DBName))

	return &DB{gorm: gormDB}, nil
}

// GetDbContext returns a request-scoped handle. It is a session over the shared
// pool, not a new connection, so controllers can call it per request.
func (db *DB) GetDbContext(ctx context.Context) *gorm.DB {
	return db.gorm.WithContext(ctx)
}

// Close drains the pool on shutdown.
func (db *DB) Close() error {
	postgresDB, err := db.gorm.DB()
	if err != nil {
		return fmt.Errorf("database: get sql handle: %w", err)
	}

	return postgresDB.Close()
}

// openConnection builds the DSN from config and opens gorm. Unexported so the
// only way to a pool is New, which owns its lifecycle.
func openConnection(appConfig config.Config) (*gorm.DB, error) {
	pgConfig := appConfig.Postgres

	gormLogger := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             slowThreshold,
			LogLevel:                  logLevel(pgConfig.Debug),
			IgnoreRecordNotFoundError: true,
			Colorful:                  !appConfig.IsProduction(),
		},
	)

	gormDB, err := gorm.Open(postgres.Open(buildDSN(pgConfig)), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("database: open %s/%s: %w", pgConfig.Host, pgConfig.DBName, err)
	}

	return gormDB, nil
}

// buildDSN is the connection string handed to the driver. Credentials are URL
// escaped, so a password with punctuation in it does not corrupt the URL.
func buildDSN(pg config.Postgres) string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(pg.Username, pg.Password),
		Host:     pg.Host,
		Path:     "/" + pg.DBName,
		RawQuery: url.Values{"sslmode": {pg.SSLMode}}.Encode(),
	}

	return dsn.String()
}

// logLevel maps app.postgres.debug onto gorm's logger: statement-level logging
// on, warnings only off.
func logLevel(debug bool) gormlogger.LogLevel {
	if debug {
		return gormlogger.Info
	}

	return gormlogger.Warn
}
