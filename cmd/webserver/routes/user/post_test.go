package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"weezel/example-gin/pkg/generated/sqlc"

	"github.com/gin-gonic/gin"
)

func TestHandlerController_PostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type fields struct {
		querier sqlc.Querier
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		fields    fields
		args      args
		postCheck func(*httptest.ResponseRecorder)
		recorder  *httptest.ResponseRecorder
		request   *http.Request
		name      string
	}{
		{
			name: "Add user with name only",
			fields: fields{
				querier: MockHandlerController{
					mAdduser: func(context.Context, sqlc.AddUserParams) (int32, error) {
						return 0, nil
					},
				},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusCreated {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusOK)
				}

				expectedBody := `{"msg":"Added user 'Pope'"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request: httptest.NewRequest(
				http.MethodPost,
				"/user",
				strings.NewReader(`{"name": "Pope"}`),
			),
		},
		{
			name: "Add user with nil payload",
			fields: fields{
				// We bail out before executing querier.AddUser(), hence empty implementation
				querier: MockHandlerController{},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusBadRequest)
				}

				expectedBody := `{"error":"EOF"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request:  httptest.NewRequest(http.MethodPost, "/user", nil),
		},
		{
			name: "Add user with bad name",
			fields: fields{
				// We bail out before executing querier.AddUser(), hence empty implementation
				querier: MockHandlerController{},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusBadRequest)
				}

				expectedBody := `{"error":"Name parameter empty or invalid"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request:  httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"name": ",."}`)),
		},
		{
			name: "Failed user addition",
			fields: fields{
				querier: MockHandlerController{
					mAdduser: func(context.Context, sqlc.AddUserParams) (int32, error) {
						return 0, errors.New("Expected failure when adding user")
					},
				},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusInternalServerError)
				}

				expectedBody := `{"error":"Failed to add user"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request: httptest.NewRequest(
				http.MethodPost,
				"/user",
				strings.NewReader(`{"name": "asdf"}`),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			tt.args.c = gin.CreateTestContextOnly(tt.recorder, r)
			tt.args.c.Request = tt.request
			r.POST("/user", func(c *gin.Context) {
				tt.args.c.Handler()(tt.args.c)
			})
			h := HandlerController{querier: tt.fields.querier}
			h.PostHandler(tt.args.c)
			tt.postCheck(tt.recorder)
		})
	}
}
