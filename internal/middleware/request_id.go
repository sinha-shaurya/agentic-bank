package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/internal/obs"
)

const (
	RequestIDHeader     = "X-Request-Id"
	correlationIDHeader = "X-Correlation-Id"
)

// validRequestID bounds a caller-supplied id so it can't bloat or garble log lines.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

// RequestID puts a correlation id on the request context and echoes it back:
// the caller's if well-formed, otherwise a fresh one.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := inboundRequestID(ctx)
		ctx.Request = ctx.Request.WithContext(obs.WithRequestID(ctx.Request.Context(), requestID))
		ctx.Header(RequestIDHeader, requestID)
		ctx.Next()
	}
}

func inboundRequestID(ctx *gin.Context) string {
	for _, header := range []string{RequestIDHeader, correlationIDHeader} {
		if value := ctx.GetHeader(header); validRequestID.MatchString(value) {
			return value
		}
	}
	return newRequestID()
}

func newRequestID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
