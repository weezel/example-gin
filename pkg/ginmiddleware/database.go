package ginmiddleware

import (
	"weezel/example-gin/pkg/db"

	"github.com/gin-gonic/gin"
)

func Postgres() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db.GetPool())
		c.Next()
	}
}
