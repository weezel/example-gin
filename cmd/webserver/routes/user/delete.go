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

func DeleteHandler(c *gin.Context) {
	ctx := context.Background()
	var err error

	var usr User
	if err = c.ShouldBindJSON(&usr); err != nil {
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

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
	if _, err = q.DeleteUser(ctx, sqlc.DeleteUserParams{
		ID:   usr.ID,
		Name: usr.Name,
	}); err != nil {
		l.Logger.Error().Err(err).Msg("user add")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusFailedDependency, gin.H{
		"msg": fmt.Sprintf("Deleted '%s'", usr.Name),
	})
}
