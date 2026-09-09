package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/sinha-shaurya/agentic-bank/internal/config"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func newTestServer(t *testing.T, health Health) *Server {
	t.Helper()
	return New(config.Config{Env: "test", Port: "8080"}, zap.NewNop(), health)
}

func do(t *testing.T, srv *Server, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	srv.Engine().ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder
}

func TestHealthRoutesPassWithNoChecks(t *testing.T) {
	srv := newTestServer(t, Health{})

	for _, path := range []string{"/healthz", "/readyz"} {
		if got := do(t, srv, http.MethodGet, path).Code; got != http.StatusOK {
			t.Errorf("GET %s = %d, want %d", path, got, http.StatusOK)
		}
	}
}

// A dependency that reports an error must surface as 503 — an instance that
// 200s while broken gets traffic routed to it.
func TestHealthRoutesReportDependencyFailure(t *testing.T) {
	failing := func(context.Context) error { return errors.New("database unreachable") }
	srv := newTestServer(t, Health{Live: failing, Ready: failing})

	for _, path := range []string{"/healthz", "/readyz"} {
		recorder := do(t, srv, http.MethodGet, path)
		if recorder.Code != http.StatusServiceUnavailable {
			t.Errorf("GET %s = %d, want %d", path, recorder.Code, http.StatusServiceUnavailable)
		}
		if body := recorder.Body.String(); !strings.Contains(body, "database unreachable") {
			t.Errorf("GET %s body = %q, want the failure reason", path, body)
		}
	}
}

func TestUnknownRouteAndMethodReturnJSON(t *testing.T) {
	srv := newTestServer(t, Health{})

	if got := do(t, srv, http.MethodGet, "/does-not-exist").Code; got != http.StatusNotFound {
		t.Errorf("unknown route = %d, want %d", got, http.StatusNotFound)
	}
	if got := do(t, srv, http.MethodPost, "/healthz").Code; got != http.StatusMethodNotAllowed {
		t.Errorf("wrong method = %d, want %d", got, http.StatusMethodNotAllowed)
	}
}
