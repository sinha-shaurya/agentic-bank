package httperr_test

import (
	stderrors "errors"
	"net/http"
	"strings"
	"testing"

	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
)

var errDriver = stderrors.New(`pq: password authentication failed for user "bank"`)

// A 5xx cause belongs in the logs, not in a body served to an untrusted caller.
func Test5xxCauseIsWithheldFromTheClient(t *testing.T) {
	httpErr := httperr.New(http.StatusInternalServerError, httperr.ErrInternalServer, errDriver)

	message := httpErr.GetErrorMessage()
	if message != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("message = %q, want the generic status text", message)
	}
	if strings.Contains(message, "password") {
		t.Errorf("client message leaks the driver error: %q", message)
	}
	if httpErr.GetErrorResponse().Message != message {
		t.Errorf("response payload disagrees with GetErrorMessage")
	}
}

// Withholding the cause from the client must not lose it for the logs.
func TestCauseSurvivesForLogsAndErrorsIs(t *testing.T) {
	httpErr := httperr.New(http.StatusInternalServerError, httperr.ErrInternalServer, errDriver)

	if !strings.Contains(httpErr.Error(), "password") {
		t.Errorf("Error() = %q, want the full cause", httpErr.Error())
	}
	if !stderrors.Is(httpErr, errDriver) {
		t.Error("errors.Is could not see through to the cause")
	}

	wrapped := stderrors.Join(stderrors.New("service: execute transfer"), httpErr)
	var recovered httperr.HttpError
	if !stderrors.As(wrapped, &recovered) {
		t.Fatal("errors.As did not recover the HttpError")
	}
	if recovered.GetStatusCode() != http.StatusInternalServerError {
		t.Errorf("recovered status = %d", recovered.GetStatusCode())
	}
}

func Test4xxCausePassesThrough(t *testing.T) {
	httpErr := httperr.New(
		http.StatusUnprocessableEntity,
		httperr.ErrValidation,
		stderrors.New("amount_paise must be a positive integer"),
	)

	if got := httpErr.GetErrorMessage(); got != "amount_paise must be a positive integer" {
		t.Errorf("message = %q", got)
	}
}

func TestNilCauseNeverPanics(t *testing.T) {
	httpErr := httperr.NewCode(httperr.ErrIdempotencyKeyConflict, nil)

	if got := httpErr.GetStatusCode(); got != http.StatusConflict {
		t.Errorf("status = %d, want 409", got)
	}
	if got := httpErr.GetErrorMessage(); got == "" {
		t.Error("message on a nil cause should fall back to the catalog, got empty")
	}
	if got := httpErr.Error(); got != string(httperr.ErrIdempotencyKeyConflict) {
		t.Errorf("Error() = %q", got)
	}
	if httpErr.Unwrap() != nil {
		t.Error("Unwrap() on a nil cause should be nil")
	}
}

// The catalog is the contract: a code must always be served as the same status.
func TestCatalogDrivesStatus(t *testing.T) {
	cases := map[httperr.ErrorCode]int{
		httperr.ErrValidation:             http.StatusBadRequest,
		httperr.ErrUnauthorized:           http.StatusUnauthorized,
		httperr.ErrForbidden:              http.StatusForbidden,
		httperr.ErrNotFound:               http.StatusNotFound,
		httperr.ErrIdempotencyKeyConflict: http.StatusConflict,
		httperr.ErrInsufficientFunds:      http.StatusUnprocessableEntity,
		httperr.ErrInternalServer:         http.StatusInternalServerError,
	}

	for code, wantStatus := range cases {
		httpErr := httperr.NewCode(code, stderrors.New("cause"))
		if got := httpErr.GetStatusCode(); got != wantStatus {
			t.Errorf("%s: status = %d, want %d", code, got, wantStatus)
		}
		if got := httpErr.GetErrorCode(); got != code {
			t.Errorf("%s: code = %q", code, got)
		}
	}
}

// Every catalog entry must be usable, or a code added later serves a 0 status.
func TestEveryCatalogEntryIsWellFormed(t *testing.T) {
	codes := []httperr.ErrorCode{
		httperr.ErrValidation,
		httperr.ErrUnauthorized,
		httperr.ErrForbidden,
		httperr.ErrNotFound,
		httperr.ErrIdempotencyKeyConflict,
		httperr.ErrInsufficientFunds,
		httperr.ErrInternalServer,
	}

	for _, code := range codes {
		definition, found := httperr.Lookup(code)
		if !found {
			t.Errorf("%s: not registered in the catalog", code)
			continue
		}
		if definition.StatusCode < 400 || definition.StatusCode > 599 {
			t.Errorf("%s: status = %d, want a 4xx or 5xx", code, definition.StatusCode)
		}
		if definition.Message == "" {
			t.Errorf("%s: default message is empty", code)
		}
	}
}

func TestExplicitStatusOverridesTheCatalog(t *testing.T) {
	httpErr := httperr.New(http.StatusTeapot, httperr.ErrValidation, stderrors.New("boom"))

	if got := httpErr.GetStatusCode(); got != http.StatusTeapot {
		t.Errorf("status = %d, want the explicit override", got)
	}
}

// An unregistered code must not be served as a success.
func TestUnknownCodeFallsBackToInternalServerError(t *testing.T) {
	httpErr := httperr.NewCode("err_not_a_real_code", stderrors.New("boom"))

	if got := httpErr.GetStatusCode(); got != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", got)
	}
	if got := httpErr.GetErrorMessage(); strings.Contains(got, "boom") {
		t.Errorf("unknown code leaked its cause: %q", got)
	}
}

// A 5xx catalog message is a static literal, so the fallback cannot leak either.
func TestCatalogMessageIsServedForA5xx(t *testing.T) {
	httpErr := httperr.NewCode(httperr.ErrInternalServer, errDriver)

	if got := httpErr.GetErrorMessage(); strings.Contains(got, "password") {
		t.Errorf("client message leaks the driver error: %q", got)
	}
}
