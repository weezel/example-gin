package client

import (
	"errors"
	"fmt"
	"math/bits"
	"math/rand/v2"
	"net"
	"net/http"
	"time"

	l "weezel/example-gin/pkg/logger"
)

type Middleware http.RoundTripper

// RetryConfig defines settings for the HTTP client retry middleware
type RetryConfig struct {
	MaxRetries   int
	MaxBaseDelay time.Duration // Initial delay before retrying
	MaxDelay     time.Duration // Maximum delay cap
}

// NewRetryMiddleware returns an HTTP client middleware that retries requests with exponential backoff and jitter
func NewRetryMiddleware(config RetryConfig) Middleware {
	if config.MaxBaseDelay <= 0 || config.MaxBaseDelay > time.Minute {
		l.Logger.Panic().Msg("Max base delay cannot be less than zero or greater than one minute")
	}
	if config.MaxDelay <= 0 || config.MaxDelay > (5*time.Minute) {
		l.Logger.Panic().Msg("Max delay cannot be less than zero or greater than five minutes")
	}
	if config.MaxRetries <= 0 || config.MaxRetries > 10 {
		l.Logger.Panic().Msg("Max delay cannot be less than zero or greater than ten")
	}

	return &retryTransport{
		rt:     http.DefaultTransport,
		config: config,
	}
}

// retryTransport implements http.RoundTripper with retry logic
type retryTransport struct {
	rt     http.RoundTripper
	config RetryConfig
}

func (r *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response
	//nolint:gosec // Integer is converted to uint so it fits
	for attempt := uint(1); attempt <= uint(r.config.MaxRetries); attempt++ {
		// Avoid side effects, such as body being consumed in the previous attempt,
		// header mutation and context propagation issues by cloning the context
		newReq := req.Clone(req.Context())

		resp, err = r.rt.RoundTrip(newReq)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			defer resp.Body.Close()
		}

		// Don't retry non-network related errors
		if err != nil && !isNetworkError(err) {
			return nil, fmt.Errorf("non-network error: %w", err)
		}

		l.Logger.Debug().Msgf("HTTP client retry attempt %d/%d for URL: %s",
			attempt,
			r.config.MaxRetries,
			req.URL.String(),
		)

		delay := expBackoffWithJitter(attempt, r.config.MaxBaseDelay, r.config.MaxDelay)
		select {
		case <-time.After(delay):
		case <-req.Context().Done():
			return nil, fmt.Errorf("timeout or context cancel: %w", req.Context().Err())
		}
	}

	return nil, fmt.Errorf("round trip: %w", err)
}

const maxShift = uint(bits.UintSize - 2)

func expBackoffWithJitter(attempt uint, base, maxDuration time.Duration) time.Duration {
	safeAttempt := min(attempt, maxShift)
	expBackoff := min(base * time.Duration(1<<safeAttempt))
	jitter := time.Duration(rand.Int64N(int64(expBackoff / 2))) //nolint:gosec // math/rand/v2 is okay
	delay := min(expBackoff+jitter, maxDuration)
	l.Logger.Debug().Msgf("Waiting for %s", delay)
	return delay
}

func isNetworkError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}
