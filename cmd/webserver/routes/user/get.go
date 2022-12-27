package user

import (
	"context"
	"net/http"
	"weezel/example-gin/pkg/generated/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func IndexHandler(c *gin.Context) {
	ctx := context.Background()

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
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
	ctx := context.Background()

	var name string
	if name = c.Param("name"); name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter 'name' was not given or was empty",
		})
		return
	}

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
	users, err := q.GetUser(ctx, name)
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
