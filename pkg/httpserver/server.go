package httpserver

import (
	"net/http"
	"os"
	"strings"
	"time"
	"weezel/example-gin/pkg/ginmiddleware"

	"github.com/gin-gonic/gin"
)

// New returns a new Gin engine with our custom configurations, like logging middleware.
// This is a general implementation that can be used in any server.
func New() *gin.Engine {
	if strings.ToLower(os.Getenv("DEBUG")) != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// Use our own logging middleware
	r.Use(ginmiddleware.DefaultStructuredLogger())
	r.Use(ginmiddleware.Postgres())
	r.Use(gin.Recovery())

	return r
}

func Config(r http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout: 60 * time.Second, // Mitigation against Slow loris attack (value from nginx)
		Addr:        ":8080",
		Handler:     r,
	}
}
