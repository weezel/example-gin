package user

import (
	"context"
	"net/http"
	"weezel/example-gin/pkg/db"
	"weezel/example-gin/pkg/generated/sqlc"

	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {
	ctx := context.Background()

	q := sqlc.New(db.GetPool())
	users, err := q.ListUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't get list of users",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"users": users,
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
