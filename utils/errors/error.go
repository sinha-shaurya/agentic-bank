package errors

type HttpError struct {
	ErrorCode int
	Error     error
}

type HttpErrorResponse struct {
	Message string `json:"message"`
}

func (httpError HttpError) GetStatusCode() int {
	return httpError.ErrorCode
}

func (httpError HttpError) GetErrorMessage() string {
	return httpError.Error.Error()
}

func (httpError HttpError) GetErrorResponse() HttpErrorResponse {
	return HttpErrorResponse{
		Message: httpError.GetErrorMessage(),
	}
}
