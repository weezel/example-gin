package client

import (
	"net/http"
	"time"

	l "weezel/example-gin/pkg/logger"
)

type Option func(h *HTTPClient)

func WithClient(client HTTPClienter) Option {
	return func(h *HTTPClient) {
		h.client = client
	}
}

func WithRetryConfig(config RetryConfig) Option {
	return func(h *HTTPClient) {
		h.retryConfig = &config
	}
}

type HTTPClienter interface {
	Do(*http.Request) (*http.Response, error)
}

type HTTPClient struct {
	retryConfig *RetryConfig
	client      HTTPClienter
}

// Do implements http.Client.Do interface.
// Having one client simplifies implementations, reduces duplicated code
// and it's possible to OTEL intsrument with less effort.
// HTTP client implements exponential backoff + jitter retry logic
// to avoid retry logic repetition in many places.
func (h *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	l.Logger.Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Msg("HTTP Client Do call")
	return h.client.Do(req) //nolint:wrapcheck // This is a wrapper method, don't complain
}

func NewDefaultClient(opts ...Option) *HTTPClient {
	cli := &HTTPClient{
		retryConfig: &RetryConfig{
			MaxRetries:   5,
			MaxBaseDelay: 100 * time.Millisecond,
			MaxDelay:     10 * time.Second,
		},
	}

	// Override defaults
	for _, opt := range opts {
		opt(cli)
	}

	// If no client was provided, create a default one
	if cli.client == nil {
		cli.client = &http.Client{
			Timeout:   time.Second * 30,
			Transport: NewRetryMiddleware(*cli.retryConfig),
		}
	}

	return cli
}
