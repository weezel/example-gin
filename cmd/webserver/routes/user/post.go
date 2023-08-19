package user

// This is intentionally equal to adding user.

import (
	"context"
	"fmt"
	"net/http"

	"weezel/example-gin/pkg/generated/sqlc"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
	l.Logger.Info().
		Str("user_name", usr.Name).
		Msg("Adding user")

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
	uid, err := q.AddUser(ctx, sqlc.AddUserParams{
		Name: usr.Name,
		Age:  usr.Age,
	})
	if err != nil {
		l.Logger.Error().Err(err).Msg("Adding user failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add user",
		})
		return
	}

	l.Logger.Info().
		Int32("user_id", uid).
		Str("user_name", usr.Name).
		Msg("Added user")
	c.JSON(http.StatusCreated, gin.H{
		"msg": fmt.Sprintf("Added user '%s'", usr.Name),
	})
}
