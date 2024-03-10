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
	"github.com/jackc/pgx/v5/pgtype"
)

func TestHandlerController_DeleteHandler(t *testing.T) {
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
			name: "Delete a singel user",
			fields: fields{
				querier: MockHandlerController{
					mDeleteUser: func(context.Context, string) (*sqlc.HomepageSchemaUser, error) {
						return &sqlc.HomepageSchemaUser{
							ID:    0,
							Name:  "Pope",
							Age:   68,
							City:  pgtype.Text{String: "Vatican", Valid: true},
							Phone: pgtype.Text{String: "666", Valid: true},
						}, nil
					},
				},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusOK)
				}

				expectedBody := `{"msg":"Deleted user 'Pope'"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request: httptest.NewRequest(
				http.MethodDelete,
				"/user",
				strings.NewReader(`{"name": "Pope"}`),
			),
		},
		{
			name: "Pass nil payload",
			fields: fields{
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
			request:  httptest.NewRequest(http.MethodDelete, "/user", nil),
		},
		{
			name: "Pass a bad name",
			fields: fields{
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
			request:  httptest.NewRequest(http.MethodDelete, "/user", strings.NewReader(`{"name":".,"}`)),
		},
		{
			name: "Failed user delete",
			fields: fields{
				querier: MockHandlerController{
					mDeleteUser: func(context.Context, string) (*sqlc.HomepageSchemaUser, error) {
						return nil, errors.New("Expected failure when deleting user")
					},
				},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusInternalServerError)
				}

				expectedBody := `{"error":"Failed to delete user"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request: httptest.NewRequest(
				http.MethodDelete,
				"/user",
				strings.NewReader(`{"name":"asdf"}`),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			r := gin.Default()
			tt.args.c = gin.CreateTestContextOnly(tt.recorder, r)
			tt.args.c.Request = tt.request
			r.DELETE("/user", func(_ *gin.Context) {
				tt.args.c.Handler()(tt.args.c)
			})
			h := HandlerController{querier: tt.fields.querier}
			h.DeleteHandler(tt.args.c)
			tt.postCheck(tt.recorder)
		})
	}
}
