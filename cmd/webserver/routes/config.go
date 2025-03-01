package routes

import (
	"weezel/example-gin/cmd/webserver/routes/user"
	"weezel/example-gin/pkg/generated/sqlc"
	httpserver "weezel/example-gin/pkg/http/server"
)

// Add routes to our web server
func AddRoutes(httpServer *httpserver.HTTPServer, queries *sqlc.Queries) {
	userRouterGroup := httpServer.NewRouterGroup("/user")
	ctrl := user.NewHandlerController(
		userRouterGroup,
		queries,
	)

	// User related functionality
	userRouterGroup.GET("/", ctrl.IndexHandler)
	userRouterGroup.GET(":name", ctrl.GetHandler)
	userRouterGroup.POST("", ctrl.PostHandler)
	userRouterGroup.DELETE("", ctrl.DeleteHandler)
}
