package routes

import (
	controllers "Tools3-Project/controller"
	"Tools3-Project/middleware"

	"github.com/gin-gonic/gin"
)

func EventRoutes(r *gin.Engine) {
	event := r.Group("/events")
	event.Use(middleware.AuthRequired())

	event.POST("/", controllers.CreateEvent)
	event.GET("/organized", controllers.GetMyOrganizedEvents)
	event.GET("/invited", controllers.GetMyInvitedEvents)
	event.GET("/accepted", controllers.GetMyAcceptedEvents)

	event.POST("/invite", controllers.InviteUser)
	event.DELETE("/:id", controllers.DeleteEvent)

	event.POST("/respond", controllers.RespondToEvent)
	event.GET("/attendees/:id", controllers.GetEventAttendees)

	event.GET("/search", controllers.SearchEvents)
}
