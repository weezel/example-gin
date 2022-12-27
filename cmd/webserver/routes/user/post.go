package user

import (
	"context"
	"fmt"
	"net/http"
	"weezel/example-gin/pkg/generated/sqlc"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

// PostHandler adds an user into a database
func PostHandler(c *gin.Context) {
	ctx := context.Background()
	var err error

	var usr User
	if err = c.ShouldBindJSON(&usr); err != nil {
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

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
	if _, err = q.AddUser(ctx, sqlc.AddUserParams{
		Name: usr.Name,
		Age:  usr.Age,
	}); err != nil {
		l.Logger.Error().Err(err).Msg("user add")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg": fmt.Sprintf("Added '%s'", usr.Name),
	})
}
