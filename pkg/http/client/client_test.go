package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultHTTPClient_Do(t *testing.T) {
	t.Helper()

	type fields struct {
		retryConfig RetryConfig
	}
	type args struct {
		req func(visitCounter *atomic.Uint32) (func(), *http.Request)
	}
	tests := []struct { //nolint:govet // Alignment doesn't matter in tests
		name           string
		fields         fields
		args           args
		want           *http.Response
		expectedVisits uint32
		visitCounter   atomic.Uint32
		wantErrMsg     string
	}{
		{
			name: "Succesful GET call",
			fields: fields{
				retryConfig: RetryConfig{
					MaxRetries:   2,
					MaxBaseDelay: time.Nanosecond,
					MaxDelay:     time.Millisecond * 10,
				},
			},
			args: args{
				req: func(visitCounter *atomic.Uint32) (func(), *http.Request) {
					ts := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, _ *http.Request) {
								visitCounter.Add(1)

								w.WriteHeader(http.StatusOK)
								fmt.Fprintf(w, "Work's immediately")
							}))

					req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
					if err != nil {
						t.Errorf("Failed to create request: %v", err)
					}

					return ts.Close, req
				},
			},
			expectedVisits: 1,
			want:           &http.Response{StatusCode: http.StatusInternalServerError},
			wantErrMsg:     "",
		},
		{
			name: "Retry exactly two times",
			fields: fields{
				retryConfig: RetryConfig{
					MaxRetries:   2,
					MaxBaseDelay: time.Nanosecond,
					MaxDelay:     time.Millisecond * 100,
				},
			},
			args: args{
				req: func(visitCounter *atomic.Uint32) (func(), *http.Request) {
					ts := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, _ *http.Request) {
								visitCounter.Add(1)

								w.WriteHeader(http.StatusInternalServerError)
								fmt.Fprintf(w, "Please retry")
							}))

					req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
					if err != nil {
						t.Errorf("Failed to create request: %v", err)
					}

					return ts.Close, req
				},
			},
			expectedVisits: 2,
			want:           &http.Response{StatusCode: http.StatusInternalServerError},
			wantErrMsg:     "",
		},
		{
			name: "Succeed on the third try",
			fields: fields{
				retryConfig: RetryConfig{
					MaxRetries:   3,
					MaxBaseDelay: time.Nanosecond,
					MaxDelay:     time.Millisecond,
				},
			},
			args: args{
				req: func(visitCounter *atomic.Uint32) (func(), *http.Request) {
					ts := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, _ *http.Request) {
								visitCounter.Add(1)

								if visitCounter.Load() == 3 {
									w.WriteHeader(http.StatusOK)
									fmt.Fprintf(w, "OK")
									return
								}

								w.WriteHeader(http.StatusInternalServerError)
								fmt.Fprintf(w, "Please retry")
							}))

					req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
					if err != nil {
						t.Errorf("Failed to create request: %v", err)
					}

					return ts.Close, req
				},
			},
			expectedVisits: 3,
			want:           &http.Response{StatusCode: http.StatusOK},
			wantErrMsg:     "",
		},
	}

	//nolint // Loop variable is reference by value in recent Go versions
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env var DEBUG to "true" to see retry attemps
			t.Cleanup(func() {
				// Catch panics
				if r := recover(); r != nil {
					if tt.wantErrMsg == "" {
						t.Fatalf("Paniced at the wrong place: %s", r)
					}
					// Expected panic happened so do nothing
				}
			})

			cli := NewDefaultClient(WithRetryConfig(tt.fields.retryConfig))

			closeTestServer, req := tt.args.req(&tt.visitCounter)
			t.Cleanup(func() {
				closeTestServer()
			})

			res, err := cli.Do(req)
			if (err != nil) && tt.wantErrMsg != "" {
				t.Errorf("DefaultHTTPClient.Do() error = %v, wantErr = %v", err, tt.wantErrMsg)
			}
			t.Cleanup(func() {
				if res != nil && res.Body != nil {
					res.Body.Close()
				}
			})

			if tt.visitCounter.Load() != tt.expectedVisits {
				t.Errorf("Expected visits to HTTP endpoint, got = %d, expected = %d",
					tt.visitCounter.Load(), tt.expectedVisits,
				)
			}
		})
	}
}
