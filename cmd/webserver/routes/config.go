package routes

import (
	"weezel/example-gin/cmd/webserver/routes/general"
	"weezel/example-gin/cmd/webserver/routes/user"

	"github.com/gin-gonic/gin"
)

// Add routes to our web server
func AddRoutes(r *gin.Engine) {
	r.GET("/health", general.HealthCheckHandler)

	// User related functionality
	r.GET("/user/", user.IndexHandler)
	r.GET("/user/:name", user.GetHandler)
	r.POST("/user", user.PostHandler)
	r.DELETE("/user", user.DeleteHandler)
}
