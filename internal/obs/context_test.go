package obs

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func observeGlobal(t *testing.T) *observer.ObservedLogs {
	core, logs := observer.New(zapcore.DebugLevel)
	restore := zap.ReplaceGlobals(zap.New(core))
	t.Cleanup(restore)
	return logs
}

func TestLoggerTagsRequestID(t *testing.T) {
	logs := observeGlobal(t)

	ctx := WithRequestID(context.Background(), "req-123")
	Logger(ctx).Info("user.registered")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("got %d log entries, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["request_id"]; got != "req-123" {
		t.Errorf("request_id = %v, want req-123", got)
	}
}

func TestLoggerWithoutRequestIDOmitsField(t *testing.T) {
	logs := observeGlobal(t)

	Logger(context.Background()).Info("startup")

	if _, found := logs.All()[0].ContextMap()["request_id"]; found {
		t.Error("request_id present on a context that never carried one")
	}
}
