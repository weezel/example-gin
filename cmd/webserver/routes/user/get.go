package user

import (
	"context"
	"net/http"

	"weezel/example-gin/pkg/generated/sqlc"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func IndexHandler(c *gin.Context) {
	ctx := context.Background()

	l.Logger.Info().Msg("Listing all users")

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		l.Logger.Error().Msg("No database stored in Gin context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
	users, err := q.ListUsers(ctx)
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

func GetHandler(c *gin.Context) {
	ctx := context.Background()

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

	p, ok := c.Keys["db"].(*pgxpool.Pool)
	if !ok {
		l.Logger.Error().Msg("No database stored in Gin context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect database",
		})
		return
	}
	q := sqlc.New(p)
	users, err := q.GetUser(ctx, name)
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
