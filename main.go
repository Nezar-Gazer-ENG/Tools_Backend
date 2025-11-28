package main

import (
	"Tools3-Project/config"
	controllers "Tools3-Project/controller"
	"Tools3-Project/routes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()
	controllers.InitUserCollection(db)
	controllers.InitEventCollection(db)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	store := cookie.NewStore([]byte("super-secret-key"))

	store.Options(sessions.Options{
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   60 * 60 * 24,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	r.Use(sessions.Sessions("mysession", store))

	// ROUTES
	routes.AuthRoutes(r)
	routes.EventRoutes(r)

	fmt.Println("✅ Server running on http://localhost:8080")
	r.Run(":8080")
}
