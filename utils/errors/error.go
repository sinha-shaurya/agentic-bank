package errors

import "net/http"

type ErrorCode string

const (
	// Request shape and auth.
	ErrValidation ErrorCode = "err_validation"
	ErrInternalServer ErrorCode = "err_internal_server"
)

type HttpError struct {
	StatusCode int
	ErrorCode  ErrorCode
	Error      error
}


func New(statusCode int, errorCode ErrorCode, err error) HttpError {
	return HttpError{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Error:      err,
	}
}

type HttpErrorResponse struct {
	Message string `json:"message"`
}

// GetStatusCode falls back to 500 rather than gin's implicit 200: an
// HttpError built without a status must never be served as a success.
func (httpError HttpError) GetStatusCode() int {
	if httpError.StatusCode == 0 {
		return http.StatusInternalServerError
	}
	return httpError.StatusCode
}

func (httpError HttpError) GetErrorCode() ErrorCode {
	if httpError.ErrorCode == "" {
		return ErrInternalServer
	}
	return httpError.ErrorCode
}

// GetErrorMessage tolerates a nil cause; a handler that returns an HttpError
// with no wrapped error should not panic the request.
func (httpError HttpError) GetErrorMessage() string {
	if httpError.Error == nil {
		return http.StatusText(httpError.GetStatusCode())
	}
	return httpError.Error.Error()
}

func (httpError HttpError) GetErrorResponse() HttpErrorResponse {
	return HttpErrorResponse{
		Message: httpError.GetErrorMessage(),
	}
}
