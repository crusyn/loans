package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/ent/user"
	"github.com/gin-gonic/gin"
)

// mock gin context
func GetTestGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}

	return ctx
}

func TestCreateUser(t *testing.T) {

	// db init
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatal().Msgf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal().Msgf("failed creating schema resources: %v", err)
	}

	h := Handler{
		Ent: client,
	}

	for _, tc := range []struct {
		name         string
		request      userBody
		expectedCode int
	}{
		{
			name: "first chris",
			request: userBody{
				Name:    "chris",
				Social:  "123-45-6789",
				Address: "1 Apple Street",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "duplicate social",
			request: userBody{
				Name:    "chris",
				Social:  "123-45-6789",
				Address: "1 Apple Street",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "duplicate name",
			request: userBody{
				Name:    "chris",
				Social:  "000-45-6780",
				Address: "1 Apple Street",
			},
			expectedCode: http.StatusOK,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := GetTestGinContext(w)

			ctx.Request.Method = "POST"
			ctx.Request.Header.Set("Content-Type", "application/json")

			jsonbytes, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatal("could not marshal user")
			}

			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))

			h.CreateUser(ctx)

			socialExists, err := h.Ent.User.Query().Where(user.Social(tc.request.Social)).Exist(ctx)
			if err != nil {
				t.Fatalf("could not get user: %v", err)
			}

			if !socialExists {
				t.Errorf("user with social %s not saved", tc.request.Social)
			}

			if w.Code != tc.expectedCode {
				t.Errorf("unexpected status code, want: %v, got: %v", tc.expectedCode, w.Code)
			}
		})
	}

}
