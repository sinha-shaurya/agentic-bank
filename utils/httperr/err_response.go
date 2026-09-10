package httperr

import (
	"errors"
	"net/http"
)

type HttpError struct {
	StatusCode int
	ErrorCode  ErrorCode
	Err        error
}

// NewCode is the constructor to reach for: the status comes from the catalog,
// so the same code cannot be served as 400 by one handler and 422 by another.
func NewCode(errorCode ErrorCode, err error) HttpError {
	return HttpError{
		ErrorCode: errorCode,
		Err:       err,
	}
}

// New overrides the catalog status for the rare endpoint that needs a different
// one for the same code.
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

// GetStatusCode prefers an explicit override, then the catalog, then 500 rather
// than gin's implicit 200: an unrecognised code must never be served as success.
func (httpError HttpError) GetStatusCode() int {
	if httpError.StatusCode != 0 {
		return httpError.StatusCode
	}
	if definition, found := definitions[httpError.ErrorCode]; found {
		return definition.StatusCode
	}
	return http.StatusInternalServerError
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
// leaks schema and infrastructure detail, so callers log Err themselves. The
// fallback is the catalog message, which is a static literal and never derived
// from the cause.
func (httpError HttpError) GetErrorMessage() string {
	statusCode := httpError.GetStatusCode()
	if httpError.Err != nil && statusCode < http.StatusInternalServerError {
		return httpError.Err.Error()
	}
	if definition, found := definitions[httpError.ErrorCode]; found && definition.Message != "" {
		return definition.Message
	}
	return http.StatusText(statusCode)
}

// GetErrorResponse renders the client-visible error payload.
func (httpError HttpError) GetErrorResponse() HttpErrorResponse {
	return HttpErrorResponse{
		Message: httpError.GetErrorMessage(),
	}
}

func GetHttpError(errCode ErrorCode) HttpError {
	// lookup definition
	definition, ok := Lookup(errCode)
	if !ok {
		// TODO: Fix this
		// return ErrInternalServer
		definition, _ = Lookup(ErrInternalServer)
	}

	return New(definition.StatusCode, errCode, errors.New(definition.Message))
}
