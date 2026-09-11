package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sinha-shaurya/agentic-bank/auth/controller"
)

func AuthRoutes(router *gin.RouterGroup) {
	router = router.Group("auth")

	// add paths
	router.POST("/register", controller.RegisterUser)
	router.POST("/retrieve", controller.RetrieveUeser)
}
