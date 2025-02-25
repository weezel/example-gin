package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// visitCounter counts how many HTTP endpoint visits are performed
type visitCounter struct {
	counter map[string]uint32
	lock    sync.RWMutex
}

func (v *visitCounter) add(serviceName string) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if _, found := v.counter[serviceName]; !found {
		v.counter[serviceName] = 1
	} else {
		v.counter[serviceName]++
	}
}

func (v *visitCounter) get(serviceName string) uint32 {
	v.lock.RLock()
	defer v.lock.RUnlock()

	if _, found := v.counter[serviceName]; !found {
		return 0
	}

	return v.counter[serviceName]
}

func TestDefaultHTTPClient_Do(t *testing.T) {
	t.Helper()

	visits := visitCounter{
		counter: map[string]uint32{},
		lock:    sync.RWMutex{},
	}

	type fields struct {
		retryConfig RetryConfig
	}
	type args struct {
		req func() (func(), *http.Request)
	}
	tests := []struct { //nolint:govet // Alignment doesn't matter in tests
		name           string
		fields         fields
		args           args
		want           *http.Response
		expectedVisits uint32
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
				req: func() (func(), *http.Request) {
					ts := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, _ *http.Request) {
								// XXX Keep in sync with the test name!
								// Cannot reference to `name` field as it's initialized
								// at the same time as this. And don't want to
								// do some hacky slices, map or any other container
								// related solutions here.
								visits.add("Succesful GET call")

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
				req: func() (func(), *http.Request) {
					ts := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, _ *http.Request) {
								// XXX Keep in sync with the test name!
								// Cannot reference to `name` field as it's initialized
								// at the same time as this. And don't want to
								// do some hacky slices, map or any other container
								// related solutions here.
								visits.add("Retry exactly two times")

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
		// {
		// 	name: "Illegal base delay",
		// 	fields: fields{
		// 		retryConfig: RetryConfig{
		// 			MaxRetries:   2,
		// 			MaxBaseDelay: 0,
		// 			MaxDelay:     time.Second,
		// 		},
		// 	},
		// 	args: args{
		// 		req: func() (func(), *http.Request) {
		// 			ts := httptest.NewServer(
		// 				http.HandlerFunc(
		// 					func(w http.ResponseWriter, r *http.Request) {
		// 						w.WriteHeader(http.StatusInternalServerError)
		// 						fmt.Fprintf(w, "please retry")
		// 						return
		// 					}))

		// 			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		// 			if err != nil {
		// 				t.Errorf("Failed to create request: %v", err)
		// 			}

		// 			return ts.Close, req
		// 		},
		// 	},
		// 	expectedVisits: 0,
		// 	want:           &http.Response{StatusCode: 500},
		// 	wantErrMsg:     "should panic",
		// },
	}
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

			closeTestServer, req := tt.args.req()
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

			if visits.get(tt.name) != tt.expectedVisits {
				t.Errorf("Expected visits to HTTP endpoint, got = %d, expected = %d",
					visits.get(tt.name), tt.expectedVisits,
				)
			}
		})
	}
}
