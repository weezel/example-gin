package ginmiddleware

import (
	"time"

	l "weezel/example-gin/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// DefaultStructuredLogger logs a gin HTTP request in JSON format. Uses the
// default logger from rs/zerolog.
func DefaultStructuredLogger() gin.HandlerFunc {
	return StructuredLogger(&log.Logger)
}

// StructuredLogger logs a gin HTTP request in JSON format. Allows to set the
// logger for testing purposes.
func StructuredLogger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Fill the param
		param := gin.LogFormatterParams{
			// Request:      &http.Request{},
			TimeStamp:    start,
			StatusCode:   c.Writer.Status(),
			Latency:      time.Since(start),
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
			BodySize:     c.Writer.Size(),
			Keys:         map[string]any{},
		}
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		if c.Request.URL.RawQuery != "" {
			param.Path = c.Request.URL.Path + "?" + c.Request.URL.RawQuery
		}

		var logEvent *zerolog.Event
		if c.Writer.Status() >= 500 {
			// Server failures are errors
			logEvent = logger.Error()
		} else if c.Writer.Status() >= 400 && c.Writer.Status() <= 499 {
			// Client failures are warnings
			logEvent = logger.Warn()
		} else {
			logEvent = logger.Info()
		}

		logEvent.Str("uniq_id", l.UniqID()).
			Str("client_ip", param.ClientIP).
			Str("method", param.Method).
			Int("status_code", param.StatusCode).
			Int("body_size", param.BodySize).
			Str("path", param.Path).
			Str("latency", param.Latency.String()).
			Msg(param.ErrorMessage)
	}
}