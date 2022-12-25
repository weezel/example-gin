package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PostHandler adds an user into a database
func PostHandler(c *gin.Context) {
	var usr User
	if err := c.ShouldBindJSON(&usr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !isValidName(usr.Name) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Not a valid name",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg": fmt.Sprintf("Added '%s'", usr.Name),
	})
}
