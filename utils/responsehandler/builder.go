package responsehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/utils/errors"
)

//nolint
type Builder struct {
	ctx        *gin.Context
	httpStatus int
	response   Response
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
//nolint
func (builder *Builder) SetResponse(data any) *Builder {
	builder.httpStatus = http.StatusOK
	// skip error code setting
	builder.response = Response{Status: StatusSuccess, Message: defaultSuccessMessage, Data: data}
	return builder
}

//nolint
func (builder *Builder) SetApiError(httpError errors.HttpError) *Builder {
	builder.httpStatus = httpError.GetStatusCode()
	builder.response = Response{Status: StatusFailure, Message: httpError.GetErrorResponse().Message, ErrorCode: httpError.GetErrorCode()}
	return builder
}

// nolint
func (builder *Builder) SetHttpStatus(httpStatus int) *Builder {
	builder.httpStatus = httpStatus
	return builder
}

//nolint
func (builder *Builder) SetMessage(message string) *Builder {
	builder.response.Message = message
	return builder
}

// Build writes the envelope. It is the only place in the service that
// serializes an API response, and must be called exactly once per request.
func (builder *Builder) Build() {
	builder.ctx.JSON(builder.httpStatus, builder.response)
}
