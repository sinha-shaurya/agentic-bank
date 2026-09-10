package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func newTestContext(method, target string, body string, params gin.Params) *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
		ctx.Request = httptest.NewRequest(method, target, reader)
		ctx.Request.Header.Set("Content-Type", "application/json")
	} else {
		ctx.Request = httptest.NewRequest(method, target, nil)
	}
	ctx.Params = params
	return ctx
}

// assertBadRequest recovers the HttpError channel and returns its message.
func assertBadRequest(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		t.Fatal("err = nil, want a validation error")
	}

	var httpErr httperr.HttpError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %#v, want an httperr.HttpError", err)
	}
	if httpErr.GetStatusCode() != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", httpErr.GetStatusCode())
	}
	if httpErr.GetErrorCode() != httperr.ErrValidation {
		t.Errorf("code = %q, want %q", httpErr.GetErrorCode(), httperr.ErrValidation)
	}
	return httpErr.GetErrorMessage()
}

// A read DTO tags its id for the uri and for the response body.
type getWalletRequest struct {
	ID string `uri:"id" json:"id" validate:"required"`
}

func TestGetWithJSONTaggedDTODoesNotBindBody(t *testing.T) {
	ctx := newTestContext(http.MethodGet, "/wallets/w_1", "", gin.Params{{Key: "id", Value: "w_1"}})

	var req getWalletRequest
	if err := BindAndValidate(ctx, &req); err != nil {
		t.Fatalf("err = %v, want nil on a valid GET", err)
	}
	if req.ID != "w_1" {
		t.Errorf("id = %q, want %q", req.ID, "w_1")
	}
}

type transferRequest struct {
	From           string `json:"from"           validate:"required"`
	To             string `json:"to"             validate:"required"`
	AmountPaise    int64  `json:"amount_paise"   validate:"gt=0"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

func TestPostBindsJSONBody(t *testing.T) {
	body := `{"from":"w_1","to":"w_2","amount_paise":150,"idempotency_key":"k1"}`
	ctx := newTestContext(http.MethodPost, "/transfers", body, nil)

	var req transferRequest
	if err := BindAndValidate(ctx, &req); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if req.AmountPaise != 150 || req.From != "w_1" || req.IdempotencyKey != "k1" {
		t.Errorf("bound = %+v", req)
	}
}

// An empty body must name the missing fields, not surface a bare io.EOF.
func TestPostWithNoBodyReportsMissingFields(t *testing.T) {
	ctx := newTestContext(http.MethodPost, "/transfers", "", nil)

	var req transferRequest
	message := assertBadRequest(t, BindAndValidate(ctx, &req))

	if strings.Contains(message, "EOF") {
		t.Errorf("message leaks the decoder error: %q", message)
	}
	if !strings.Contains(message, "from is required") {
		t.Errorf("message = %q, want it to name the missing fields", message)
	}
}

func TestMalformedJSONIsCleanBadRequest(t *testing.T) {
	ctx := newTestContext(http.MethodPost, "/transfers", `{"from":`, nil)

	var req transferRequest
	assertBadRequest(t, BindAndValidate(ctx, &req))
}

// The message names the json field the client sent and the tag that failed.
func TestValidationMessageUsesJSONNameAndTag(t *testing.T) {
	body := `{"from":"w_1","to":"w_2","amount_paise":0,"idempotency_key":"k1"}`
	ctx := newTestContext(http.MethodPost, "/transfers", body, nil)

	var req transferRequest
	message := assertBadRequest(t, BindAndValidate(ctx, &req))

	if message != "amount_paise must be greater than 0" {
		t.Errorf("message = %q", message)
	}
}

func TestMultipleFailuresAreAllReported(t *testing.T) {
	ctx := newTestContext(http.MethodPost, "/transfers", `{"from":"w_1"}`, nil)

	var req transferRequest
	message := assertBadRequest(t, BindAndValidate(ctx, &req))

	for _, want := range []string{"to is required", "amount_paise must be greater than 0", "idempotency_key is required"} {
		if !strings.Contains(message, want) {
			t.Errorf("message = %q, missing %q", message, want)
		}
	}
}
