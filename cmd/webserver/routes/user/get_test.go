package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"weezel/example-gin/pkg/generated/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestHandlerController_IndexHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type jsonResponse struct {
		Users []*sqlc.HomepageSchemaUser `json:"users"`
	}

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
		name      string
	}{
		{
			name: "List zero users",
			fields: fields{
				querier: MockHandlerController{
					mListUsers: func(context.Context) ([]*sqlc.HomepageSchemaUser, error) {
						return []*sqlc.HomepageSchemaUser{}, nil
					},
				},
			},
			recorder: httptest.NewRecorder(),
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusOK)
				}

				expectedBody := "{\n    \"users\": []\n}"
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
		},
		{
			name: "List a couple of users",
			fields: fields{
				querier: MockHandlerController{
					mListUsers: func(context.Context) ([]*sqlc.HomepageSchemaUser, error) {
						return []*sqlc.HomepageSchemaUser{
							{
								ID:    0,
								Name:  "Esko",
								Age:   12,
								City:  pgtype.Text{String: "Amsterdam", Valid: true},
								Phone: pgtype.Text{String: "123-call", Valid: true},
							},
							{
								ID:    1,
								Name:  "Pesko",
								Age:   99,
								City:  pgtype.Text{String: "London", Valid: true},
								Phone: pgtype.Text{String: "999-call", Valid: true},
							},
						}, nil
					},
				},
			},
			recorder: httptest.NewRecorder(),
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusOK)
				}

				expectedOutput := []*sqlc.HomepageSchemaUser{
					{
						ID:    0,
						Name:  "Esko",
						Age:   12,
						City:  pgtype.Text{String: "Amsterdam", Valid: true},
						Phone: pgtype.Text{String: "123-call", Valid: true},
					},
					{
						ID:    1,
						Name:  "Pesko",
						Age:   99,
						City:  pgtype.Text{String: "London", Valid: true},
						Phone: pgtype.Text{String: "999-call", Valid: true},
					},
				}
				expectedResponseBody, err := json.MarshalIndent(
					jsonResponse{Users: expectedOutput},
					"",
					"    ",
				)
				if err != nil {
					t.Error(err)
				}
				responseBody := recorder.Body.String()
				if diff := cmp.Diff(string(expectedResponseBody), responseBody); diff != "" {
					t.Errorf("Response body differs from expected:\n%s\n", diff)
				}
			},
		},
		{
			name: "Erroneous ListUsers() call",
			fields: fields{
				querier: MockHandlerController{
					mListUsers: func(context.Context) ([]*sqlc.HomepageSchemaUser, error) {
						return nil, errors.New("ListUsers failed as expected")
					},
				},
			},
			recorder: httptest.NewRecorder(),
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusInternalServerError)
				}

				expectedBody := `{"error":"Couldn't get list of users"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s",
						responseBody, expectedBody)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Hook up recorder to a new Gin context here
			tt.args.c = gin.CreateTestContextOnly(tt.recorder, gin.Default())
			h := HandlerController{querier: tt.fields.querier}
			h.IndexHandler(tt.args.c)
			tt.postCheck(tt.recorder)
		})
	}
}

func TestHandlerController_GetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exampleGandalf := &sqlc.HomepageSchemaUser{
		ID:    0,
		Name:  "Gandalf",
		Age:   990,
		City:  pgtype.Text{String: "Tokyo", Valid: true},
		Phone: pgtype.Text{String: "+99223828", Valid: true},
	}

	type jsonResponse struct {
		Users *sqlc.HomepageSchemaUser `json:"users"`
	}

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
		param     gin.Param
		name      string
	}{
		{
			name: "Get a user",
			fields: fields{
				querier: MockHandlerController{
					mGetUser: func(context.Context, string) (*sqlc.HomepageSchemaUser, error) {
						return exampleGandalf, nil
					},
				},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusOK)
				}

				expectedBody, err := json.MarshalIndent(jsonResponse{Users: exampleGandalf}, "", "    ")
				if err != nil {
					t.Error(err)
				}
				responseBody := recorder.Body.String()
				if responseBody != string(expectedBody) {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request:  httptest.NewRequest(http.MethodGet, "/user/Gandalf", nil),
			param:    gin.Param{Key: "name", Value: "Gandalf"},
		},
		{
			name: "User with malformed name",
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
			request:  httptest.NewRequest(http.MethodGet, "/user/.,", nil),
			param:    gin.Param{Key: "name", Value: ".,"},
		},
		{
			name: "Get a user from database fails",
			fields: fields{
				querier: MockHandlerController{
					mGetUser: func(context.Context, string) (*sqlc.HomepageSchemaUser, error) {
						return nil, errors.New("expected error when getting user from DB")
					},
				},
			},
			postCheck: func(recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("Wrong status code returned, got=%d, want=%d",
						recorder.Code, http.StatusInternalServerError)
				}

				expectedBody := `{"error":"Couldn't get list of users"}`
				responseBody := recorder.Body.String()
				if responseBody != expectedBody {
					t.Errorf("Response body differs, got=%s, want=%s", responseBody, expectedBody)
				}
			},
			recorder: httptest.NewRecorder(),
			request:  httptest.NewRequest(http.MethodGet, "/user/Gandalf", nil),
			param:    gin.Param{Key: "name", Value: "Gandalf"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Hook up recorder to a new Gin context here
			r := gin.Default()
			tt.args.c = gin.CreateTestContextOnly(tt.recorder, r)
			tt.args.c.Request = tt.request
			tt.args.c.Params = append(tt.args.c.Params, tt.param)
			r.GET("/user/:name", func(_ *gin.Context) {
				tt.args.c.Handler()(tt.args.c)
			})
			h := HandlerController{querier: tt.fields.querier}
			h.GetHandler(tt.args.c)
			tt.postCheck(tt.recorder)
		})
	}
}
