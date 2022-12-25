package general

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheckHandler(c *gin.Context) {
	// TODO disable logging
	c.String(http.StatusOK, "OK\n")
}
