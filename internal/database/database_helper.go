// utility package for database setup
package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
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

// pool is the single shared connection pool, set up once by New.
var (
	pool     *gorm.DB
	initOnce sync.Once
	initErr  error
)

// New initialises the shared pool. Repeat calls reuse it instead of opening another.
func New(appConfig config.Config, logger *zap.Logger) error {
	initOnce.Do(func() {
		initErr = initPool(appConfig, logger)
	})

	return initErr
}

func initPool(appConfig config.Config, logger *zap.Logger) error {
	gormDB, err := openConnection(appConfig)
	if err != nil {
		return err
	}

	postgresDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("database: get sql handle: %w", err)
	}

	// set connection properties
	postgresDB.SetMaxIdleConns(maxIdleConns)
	postgresDB.SetMaxOpenConns(maxOpenConns)
	postgresDB.SetConnMaxLifetime(connMaxLifetime)

	logger.Info("db context loaded successfully",
		zap.String("host", appConfig.Postgres.Host),
		zap.String("dbname", appConfig.Postgres.DBName))

	pool = gormDB
	return nil
}

// GetDbContext returns a request-scoped session over the shared pool, not a new connection.
func GetDbContext(ctx context.Context) *gorm.DB {
	if pool == nil {
		panic("database: GetDbContext called before New")
	}

	return pool.WithContext(ctx)
}

// Close drains the shared pool on shutdown.
func Close() error {
	if pool == nil {
		return nil
	}

	postgresDB, err := pool.DB()
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
