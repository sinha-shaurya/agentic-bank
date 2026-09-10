// Package errors defines the service→controller error channel. Importers that
// also need the standard library's errors package alias this one as apierrors.
package errors

import "net/http"

// ErrorCode is the machine-readable discriminator on the wire.
type ErrorCode string

const (
	// Request shape and auth.
	ErrValidation     ErrorCode = "err_validation"
	ErrInternalServer ErrorCode = "err_internal_server"
)

// HttpError carries an HTTP status, a wire error code, and the cause. It
// implements error and unwraps to Err.
type HttpError struct {
	StatusCode int
	ErrorCode  ErrorCode
	Err        error
}

// New builds an HttpError.
func New(statusCode int, errorCode ErrorCode, err error) HttpError {
	return HttpError{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Err:        err,
	}
}

// HttpErrorResponse is the error payload shape.
type HttpErrorResponse struct {
	Message string `json:"message"`
}

// Error reports the full cause, for logs and errors.As. The response body uses
// GetErrorMessage instead.
func (httpError HttpError) Error() string {
	if httpError.Err == nil {
		return string(httpError.GetErrorCode())
	}
	return httpError.Err.Error()
}

// Unwrap exposes the cause to errors.Is and errors.As.
func (httpError HttpError) Unwrap() error {
	return httpError.Err
}

// GetStatusCode falls back to 500 rather than gin's implicit 200: an
// HttpError built without a status must never be served as a success.
func (httpError HttpError) GetStatusCode() int {
	if httpError.StatusCode == 0 {
		return http.StatusInternalServerError
	}
	return httpError.StatusCode
}

// GetErrorCode falls back to the internal-server code so an error response
// always carries a discriminator.
func (httpError HttpError) GetErrorCode() ErrorCode {
	if httpError.ErrorCode == "" {
		return ErrInternalServer
	}
	return httpError.ErrorCode
}

// GetErrorMessage is the client-visible message. A 5xx cause is withheld — it
// leaks schema and infrastructure detail, so callers log Err themselves — and a
// nil cause does not panic.
func (httpError HttpError) GetErrorMessage() string {
	statusCode := httpError.GetStatusCode()
	if httpError.Err == nil || statusCode >= http.StatusInternalServerError {
		return http.StatusText(statusCode)
	}
	return httpError.Err.Error()
}

// GetErrorResponse renders the client-visible error payload.
func (httpError HttpError) GetErrorResponse() HttpErrorResponse {
	return HttpErrorResponse{
		Message: httpError.GetErrorMessage(),
	}
}
