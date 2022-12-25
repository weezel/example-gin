package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {
	// TODO
	c.JSON(http.StatusOK, gin.H{
		"name": "TODO",
	})
}

func GetHandler(c *gin.Context) {
	var name string
	if name = c.Param("name"); name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter 'name' was not given or was empty",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name": name,
	})
}
