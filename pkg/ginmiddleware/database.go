package ginmiddleware

import (
	"weezel/example-gin/pkg/postgres"

	"github.com/gin-gonic/gin"
)

func Postgres() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", postgres.GetPool())
		c.Next()
	}
}
