package user

// This is intentionally equal to deleting user.

import (
	"context"
	"fmt"
	"net/http"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
)

func (h HandlerController) DeleteHandler(c *gin.Context) {
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

	if _, err = h.querier.DeleteUser(ctx, usr.Name); err != nil {
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
