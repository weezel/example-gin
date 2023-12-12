package routes

import (
	"weezel/example-gin/cmd/webserver/routes/user"
	"weezel/example-gin/pkg/generated/sqlc"
	"weezel/example-gin/pkg/httpserver"
)

// Add routes to our web server
func AddRoutes(httpServer *httpserver.HTTPServer, queries *sqlc.Queries) {
	ctrl := user.NewHandlerController(queries) // TODO

	// User related functionality
	httpServer.GET("/user/", ctrl.IndexHandler)
	httpServer.GET("/user/:name", ctrl.GetHandler)
	httpServer.POST("/user", ctrl.PostHandler)
	httpServer.DELETE("/user", ctrl.DeleteHandler)
}
