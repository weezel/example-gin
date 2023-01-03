package general

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheckHandler(c *gin.Context) {
	// This route isn't logged. See our custom logging middleware for more info.
	c.String(http.StatusOK, "OK\n")
}
