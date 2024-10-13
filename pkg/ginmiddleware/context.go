package ginmiddleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func ContextMiddleware(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("ctx", ctx)
		c.Next()
	}
}
