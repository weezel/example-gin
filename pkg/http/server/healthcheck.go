package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// This function intentionally has package visibility. Function WithCustomHealthCheckHandler()
// must be used for health check customizations.
func healthCheckHandler(c *gin.Context) {
	// This route isn't logged. See our custom logging middleware for more info.
	c.String(http.StatusOK, "OK\n")
}
