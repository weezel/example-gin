package user

import (
	"net/http"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
)

func (h HandlerController) IndexHandler(c *gin.Context) {
	ctx := c.Request.Context()

	l.Logger.Info().Msg("Listing all users")

	users, err := h.querier.ListUsers(ctx)
	if err != nil {
		l.Logger.Error().Err(err).Msg("Listing all users failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't get list of users",
		})
		return
	}

	l.Logger.Info().Msg("Listed all users")
	c.IndentedJSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func (h HandlerController) GetHandler(c *gin.Context) {
	ctx := c.Request.Context()

	name := c.Param("name")
	l.Logger.Info().
		Str("user_name", name).
		Msg("Getting user")
	if name == "" || !isValidName(name) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name parameter empty or invalid",
		})
		return
	}

	users, err := h.querier.GetUser(ctx, name)
	if err != nil {
		l.Logger.Info().Err(err).Msg("Getting user failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't get list of users",
		})
		return
	}

	l.Logger.Info().
		Str("user_name", name).
		Msg("Got user")
	c.IndentedJSON(http.StatusOK, gin.H{
		"users": users,
	})
}
