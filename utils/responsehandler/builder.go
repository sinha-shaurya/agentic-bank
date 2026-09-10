package responsehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
)

// Builder accumulates one response for one request. Setters mutate individual
// fields, so they compose in any order.
type Builder struct {
	ctx        *gin.Context
	httpStatus int
	response   Response

	// messageOverridden keeps an explicit SetMessage from being overwritten by
	// a later setter's default.
	messageOverridden bool
}

// New starts a response for this request. The zero state is a well-formed 200
// success, so a handler that reaches Build without calling a setter still emits
// a valid envelope rather than writing HTTP status 0 and panicking.
func New(ctx *gin.Context) *Builder {
	return &Builder{
		ctx:        ctx,
		httpStatus: http.StatusOK,
		response: Response{
			Status:  StatusSuccess,
			Message: defaultSuccessMessage,
		},
	}
}

// SetResponse marks the response a success and attaches its payload, clearing
// any error state set earlier.
func (builder *Builder) SetResponse(data any) *Builder {
	builder.httpStatus = http.StatusOK
	builder.response.Status = StatusSuccess
	builder.response.Data = data
	builder.response.ErrorCode = ""
	if !builder.messageOverridden {
		builder.response.Message = defaultSuccessMessage
	}
	return builder
}

// SetApiError marks the response a failure, taking the status and wire error
// code from the HttpError and clearing any payload set earlier.
func (builder *Builder) SetApiError(httpError httperr.HttpError) *Builder {
	builder.httpStatus = httpError.GetStatusCode()
	builder.response.Status = StatusFailure
	builder.response.Data = nil
	builder.response.ErrorCode = httpError.GetErrorCode()
	if !builder.messageOverridden {
		builder.response.Message = httpError.GetErrorResponse().Message
	}
	return builder
}

// SetHttpStatus overrides the HTTP status for successes that are not 200.
func (builder *Builder) SetHttpStatus(httpStatus int) *Builder {
	builder.httpStatus = httpStatus
	return builder
}

// SetMessage overrides the human-readable message, in any order.
func (builder *Builder) SetMessage(message string) *Builder {
	builder.response.Message = message
	builder.messageOverridden = true
	return builder
}

// Build writes the envelope. It is the only place in the service that
// serializes an API response, and must be called exactly once per request.
func (builder *Builder) Build() {
	builder.ctx.JSON(builder.httpStatus, builder.response)
}
