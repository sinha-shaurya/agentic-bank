package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sinha-shaurya/agentic-bank/auth/service"
	"github.com/sinha-shaurya/agentic-bank/internal/database"
	"github.com/sinha-shaurya/agentic-bank/utils/responsehandler"
)

// RegisterUser mints a new caller identity. Build is called per branch, not
// deferred, so a panic can't flush the builder's default 200 before Recovery.
func RegisterUser(ctx *gin.Context) {
	responseBuilder := responsehandler.New(ctx)

	db := database.GetDbContext(ctx)
	response, apiErr := service.RegisterUser(ctx, db)
	if apiErr != nil {
		responseBuilder.SetApiError(*apiErr).Build()
		return
	}

	responseBuilder.SetResponse(response).SetHttpStatus(http.StatusCreated).Build()
}
