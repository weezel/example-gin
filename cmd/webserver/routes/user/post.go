package user

import (
	"context"
	"fmt"
	"net/http"
	"weezel/example-gin/pkg/db"
	"weezel/example-gin/pkg/generated/sqlc"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
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

	q := sqlc.New(db.GetPool())
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
