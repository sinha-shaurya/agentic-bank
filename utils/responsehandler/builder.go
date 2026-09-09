package responsehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/utils/errors"
)


type Builder struct {
	ctx        *gin.Context
	httpStatus int
	status     int
	message    string
	data       any
	errorCode  errors.ErrorCode
}

func New(ctx *gin.Context) *Builder {
	return &Builder{
		ctx:        ctx,
		httpStatus: http.StatusOK,
		status:     StatusSuccess,
		message:    defaultSuccessMessage,
	}
}


func (builder *Builder) SetResponse(data any) *Builder {
	builder.status = StatusSuccess
	builder.httpStatus = http.StatusOK
	builder.message = defaultSuccessMessage
	builder.data = data
	builder.errorCode = ""
	return builder
}


func (builder *Builder) SetApiError(httpError errors.HttpError) *Builder {
	builder.status = StatusFailure
	builder.httpStatus = httpError.GetStatusCode()
	builder.message = httpError.GetErrorMessage()
	builder.data = httpError.GetErrorResponse()
	builder.errorCode = httpError.GetErrorCode()
	return builder
}


func (builder *Builder) SetHttpStatus(httpStatus int) *Builder {
	builder.httpStatus = httpStatus
	return builder
}


func (builder *Builder) SetMessage(message string) *Builder {
	builder.message = message
	return builder
}


func (builder *Builder) Build() {
	builder.ctx.JSON(builder.httpStatus, Response{
		Status:    builder.status,
		Message:   builder.message,
		Data:      builder.data,
		ErrorCode: builder.errorCode,
	})
}
