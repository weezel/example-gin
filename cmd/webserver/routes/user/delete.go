package user

// This is intentionally equal to deleting user.

import (
	"context"
	"fmt"
	"net/http"

	"weezel/example-gin/pkg/generated/sqlc"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
	l.Logger.Info().
		Int32("user_id", usr.ID).
		Str("user_name", usr.Name).
		Msg("Deleting user")

	if usr.Name == "" || !isValidName(usr.Name) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name parameter empty or invalid",
		})
		return
	}

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		l.Logger.Error().Msg("No database stored in Gin context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
	if _, err = q.DeleteUser(ctx, usr.Name); err != nil {
		l.Logger.Info().Err(err).Msg("Deleting user failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	l.Logger.Info().
		Int32("user_id", usr.ID).
		Str("user_name", usr.Name).
		Msg("Deleted user")
	c.JSON(http.StatusOK, gin.H{
		"msg": fmt.Sprintf("Deleted user '%s'", usr.Name),
	})
}
