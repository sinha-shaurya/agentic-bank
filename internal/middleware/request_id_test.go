package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/internal/obs"
)

var generatedID = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// newRouter mirrors main.go: the handler hands *gin.Context to obs, exactly as
// controllers hand it to the service layer.
func newRouter() *gin.Engine {
	router := gin.New()
	router.ContextWithFallback = true
	router.Use(RequestID())
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, obs.RequestID(ctx))
	})
	return router
}

func serve(headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	newRouter().ServeHTTP(recorder, req)
	return recorder
}

func TestRequestIDResolution(t *testing.T) {
	for _, tc := range []struct {
		name    string
		headers map[string]string
		want    string // empty means a generated id is expected
	}{
		{"generated when absent", nil, ""},
		{"honours X-Request-Id", map[string]string{RequestIDHeader: "req-abc.123"}, "req-abc.123"},
		{"falls back to X-Correlation-Id", map[string]string{correlationIDHeader: "corr-1"}, "corr-1"},
		{"X-Request-Id wins", map[string]string{RequestIDHeader: "req-1", correlationIDHeader: "corr-1"}, "req-1"},
		{"replaces malformed", map[string]string{RequestIDHeader: "bad id\n{}"}, ""},
		{"replaces oversized", map[string]string{RequestIDHeader: strings.Repeat("a", 129)}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := serve(tc.headers)

			seen := recorder.Body.String()
			if echoed := recorder.Header().Get(RequestIDHeader); echoed != seen {
				t.Errorf("response header %q != id seen by handler %q", echoed, seen)
			}
			if tc.want == "" && !generatedID.MatchString(seen) {
				t.Errorf("id = %q, want a generated 32-char hex id", seen)
			}
			if tc.want != "" && seen != tc.want {
				t.Errorf("id = %q, want %q", seen, tc.want)
			}
		})
	}
}

func TestGeneratedRequestIDsAreUnique(t *testing.T) {
	first, second := serve(nil).Body.String(), serve(nil).Body.String()
	if first == second {
		t.Errorf("two requests shared id %q", first)
	}
}
