package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// checkTimeout bounds a health probe so a hung dependency surfaces as a failed
// check rather than a hung request.
const checkTimeout = 3 * time.Second

// Check reports whether one dependency is usable right now.
type Check func(ctx context.Context) error

// Health carries the probes behind /healthz and /readyz.
//
// Live answers "is this process working" — it is what the container runtime and
// Render's health check poll. Ready answers "may this instance take traffic",
// which additionally covers one-time startup work such as migrations.
//
// A nil check means "nothing to verify yet" and passes. It is never used to
// paper over a dependency that reported an error: a failing check returns 503,
// so an orchestrator does not route traffic to a broken instance.
type Health struct {
	Live  Check
	Ready Check
}

type healthResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func registerHealthRoutes(engine *gin.Engine, health Health) {
	engine.GET("/healthz", healthHandler(health.Live))
	engine.GET("/readyz", healthHandler(health.Ready))
}

func healthHandler(check Check) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if check == nil {
			ctx.JSON(http.StatusOK, healthResponse{Status: "ok"})
			return
		}

		checkCtx, cancel := context.WithTimeout(ctx.Request.Context(), checkTimeout)
		defer cancel()

		if err := check(checkCtx); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, healthResponse{
				Status: "unavailable",
				Reason: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, healthResponse{Status: "ok"})
	}
}
