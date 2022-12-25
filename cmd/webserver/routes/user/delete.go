package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeleteHandler(c *gin.Context) {
	var usr User
	if err := c.ShouldBindJSON(&usr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if usr.Name == "" || !isValidName(usr.Name) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter name empty or invalid",
		})
		return
	}
	c.JSON(http.StatusFailedDependency, gin.H{
		"msg": fmt.Sprintf("Deleted '%s'", usr.Name),
	})
}
