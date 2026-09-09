package obs

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds the single process-wide logger. It is created once at
// startup and injected — never constructed per request, which would allocate a
// logger (and, in the pattern this replaces, a database connection) on every
// call.
//
// Output is JSON in production so log lines are machine-readable wherever they
// are shipped, and the friendlier console encoder locally.
func NewLogger(level string, production bool) (*zap.Logger, error) {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("obs: parse log level %q: %w", level, err)
	}

	var cfg zap.Config
	if production {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(parsedLevel)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("obs: build logger: %w", err)
	}

	return logger, nil
}
