package errors_test

import (
	stderrors "errors"
	"net/http"
	"strings"
	"testing"

	apierrors "github.com/sinha-shaurya/agentic-bank/utils/errors"
)

var errDriver = stderrors.New(`pq: password authentication failed for user "bank"`)

// A 5xx cause belongs in the logs, not in a body served to an untrusted caller.
func Test5xxCauseIsWithheldFromTheClient(t *testing.T) {
	httpErr := apierrors.New(http.StatusInternalServerError, apierrors.ErrInternalServer, errDriver)

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
	httpErr := apierrors.New(http.StatusInternalServerError, apierrors.ErrInternalServer, errDriver)

	if !strings.Contains(httpErr.Error(), "password") {
		t.Errorf("Error() = %q, want the full cause", httpErr.Error())
	}
	if !stderrors.Is(httpErr, errDriver) {
		t.Error("errors.Is could not see through to the cause")
	}

	wrapped := stderrors.Join(stderrors.New("service: execute transfer"), httpErr)
	var recovered apierrors.HttpError
	if !stderrors.As(wrapped, &recovered) {
		t.Fatal("errors.As did not recover the HttpError")
	}
	if recovered.GetStatusCode() != http.StatusInternalServerError {
		t.Errorf("recovered status = %d", recovered.GetStatusCode())
	}
}

func Test4xxCausePassesThrough(t *testing.T) {
	httpErr := apierrors.New(
		http.StatusUnprocessableEntity,
		apierrors.ErrValidation,
		stderrors.New("amount_paise must be a positive integer"),
	)

	if got := httpErr.GetErrorMessage(); got != "amount_paise must be a positive integer" {
		t.Errorf("message = %q", got)
	}
}

func TestNilCauseNeverPanics(t *testing.T) {
	httpErr := apierrors.New(http.StatusConflict, apierrors.ErrValidation, nil)

	if got := httpErr.GetErrorMessage(); got != http.StatusText(http.StatusConflict) {
		t.Errorf("message = %q", got)
	}
	if got := httpErr.Error(); got != string(apierrors.ErrValidation) {
		t.Errorf("Error() = %q", got)
	}
	if httpErr.Unwrap() != nil {
		t.Error("Unwrap() on a nil cause should be nil")
	}
}
