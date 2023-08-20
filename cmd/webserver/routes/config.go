package routes

import (
	"weezel/example-gin/cmd/webserver/routes/general"
	"weezel/example-gin/cmd/webserver/routes/user"
	"weezel/example-gin/pkg/generated/sqlc"

	"github.com/gin-gonic/gin"
)

// Add routes to our web server
func AddRoutes(r *gin.Engine, queries *sqlc.Queries) {
	ctrl := user.NewHandlerController(queries) // TODO
	r.GET("/health", general.HealthCheckHandler)

	// User related functionality
	r.GET("/user/", ctrl.IndexHandler)
	r.GET("/user/:name", ctrl.GetHandler)
	r.POST("/user", ctrl.PostHandler)
	r.DELETE("/user", ctrl.DeleteHandler)
}
