package routes

import (
	controllers "Tools3-Project/controller"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/register", controllers.Register)
	router.POST("/login", controllers.Login)
	router.GET("/logout", controllers.Logout)


}