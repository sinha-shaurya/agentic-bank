package responsehandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// newTestContext returns a gin context wired to a recorder, the way a handler
// would receive one.
func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return ctx, recorder
}

func TestSetResponseWritesSuccessEnvelope(t *testing.T) {
	ctx, recorder := newTestContext()

	New(ctx).SetResponse(gin.H{"balance_paise": int64(150)}).Build()

	if recorder.Code != http.StatusOK {
		t.Fatalf("http status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Status != StatusSuccess {
		t.Errorf("status = %d, want %d", got.Status, StatusSuccess)
	}
	if got.ErrorCode != "" {
		t.Errorf("error_code = %q, want empty on success", got.ErrorCode)
	}

	// The payload must actually reach the wire — the guard against unexported
	// envelope fields, which marshal to nothing.
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data = %#v, want an object", got.Data)
	}
	if data["balance_paise"] != float64(150) {
		t.Errorf("data.balance_paise = %v", data["balance_paise"])
	}

	// error_code must be absent from the wire on a success, not present-and-empty.
	if strings.Contains(recorder.Body.String(), "error_code") {
		t.Errorf("success body carries error_code: %s", recorder.Body.String())
	}
}

func TestSetApiErrorWritesFailureEnvelope(t *testing.T) {
	ctx, recorder := newTestContext()

	httpError := httperr.New(
		http.StatusBadRequest,
		httperr.ErrValidation,
		errors.New("amount_paise must be a positive integer"),
	)
	New(ctx).SetApiError(httpError).Build()

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("http status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Status != StatusFailure {
		t.Errorf("status = %d, want %d", got.Status, StatusFailure)
	}
	if got.ErrorCode != httperr.ErrValidation {
		t.Errorf("error_code = %q, want %q", got.ErrorCode, httperr.ErrValidation)
	}
	if got.Message != "amount_paise must be a positive integer" {
		t.Errorf("message = %q", got.Message)
	}

	// An error carries no payload: the cause is in message, the machine-readable
	// discriminator in error_code, and data is omitted from the wire entirely.
	if got.Data != nil {
		t.Errorf("data = %#v, want nil on a failure", got.Data)
	}
	if strings.Contains(recorder.Body.String(), "\"data\"") {
		t.Errorf("failure body carries data: %s", recorder.Body.String())
	}
}

// The failure flag is the zero value, so an omitempty tag on Status would drop
// it from every error response. This pins the tag.
func TestFailureStatusSurvivesSerialization(t *testing.T) {
	ctx, recorder := newTestContext()

	New(ctx).SetApiError(httperr.New(
		http.StatusInternalServerError,
		httperr.ErrInternalServer,
		errors.New("unexpected failure"),
	)).Build()

	if !strings.Contains(recorder.Body.String(), `"status":0`) {
		t.Errorf("failure body dropped status: %s", recorder.Body.String())
	}
}

func TestHttpErrorZeroValuesFallBack(t *testing.T) {
	var zero httperr.HttpError

	if got := zero.GetStatusCode(); got != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", got)
	}
	if got := zero.GetErrorCode(); got != httperr.ErrInternalServer {
		t.Errorf("code = %q, want %q", got, httperr.ErrInternalServer)
	}
	// A nil cause must not panic.
	if got := zero.GetErrorMessage(); got != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("message = %q", got)
	}
}

func TestSetResponseClearsPriorError(t *testing.T) {
	ctx, recorder := newTestContext()

	New(ctx).
		SetApiError(httperr.New(http.StatusBadRequest, httperr.ErrValidation, errors.New("boom"))).
		SetResponse(gin.H{"wallet_id": "w_1"}).
		Build()

	if recorder.Code != http.StatusOK {
		t.Fatalf("http status = %d, want 200", recorder.Code)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Status != StatusSuccess || got.ErrorCode != "" {
		t.Errorf("stale error survived: %+v", got)
	}
}

func TestSetHttpStatusOverridesSuccessStatus(t *testing.T) {
	ctx, recorder := newTestContext()

	New(ctx).
		SetResponse(gin.H{"wallet_id": "w_1"}).
		SetHttpStatus(http.StatusCreated).
		SetMessage("wallet created").
		Build()

	if recorder.Code != http.StatusCreated {
		t.Fatalf("http status = %d, want 201", recorder.Code)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Status != StatusSuccess {
		t.Errorf("status = %d, want %d", got.Status, StatusSuccess)
	}
	if got.Message != "wallet created" {
		t.Errorf("message = %q", got.Message)
	}
}

// Setters compose in any order: an explicit message survives a later setter.
func TestSetMessageWinsRegardlessOfOrder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(*Builder)
	}{
		{"message first", func(b *Builder) { b.SetMessage("wallet created").SetResponse(gin.H{"wallet_id": "w_1"}) }},
		{"message last", func(b *Builder) { b.SetResponse(gin.H{"wallet_id": "w_1"}).SetMessage("wallet created") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, recorder := newTestContext()

			builder := New(ctx)
			tc.build(builder)
			builder.Build()

			var got Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}
			if got.Message != "wallet created" {
				t.Errorf("message = %q, want %q", got.Message, "wallet created")
			}
			if got.Status != StatusSuccess || got.Data == nil {
				t.Errorf("envelope = %+v", got)
			}
		})
	}
}

// A failure must not carry a payload from an earlier success branch.
func TestSetApiErrorClearsPriorData(t *testing.T) {
	ctx, recorder := newTestContext()

	New(ctx).
		SetResponse(gin.H{"wallet_id": "w_1"}).
		SetApiError(httperr.New(http.StatusConflict, httperr.ErrValidation, errors.New("idempotency key reused"))).
		Build()

	if recorder.Code != http.StatusConflict {
		t.Fatalf("http status = %d, want 409", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "wallet_id") {
		t.Errorf("failure body carries stale data: %s", recorder.Body.String())
	}
}
