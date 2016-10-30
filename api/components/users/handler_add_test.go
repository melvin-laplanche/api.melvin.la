package users_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nivl/api.melvin.la/api/app/testhelpers"
	"github.com/Nivl/api.melvin.la/api/auth"
	"github.com/Nivl/api.melvin.la/api/components/users"
	"github.com/stretchr/testify/assert"
)

func TestHandlerAdd(t *testing.T) {
	defer testhelpers.PurgeModels(t)

	tests := []struct {
		description string
		code        int
		params      *users.HandlerAddParams
	}{
		{
			"Empty User",
			http.StatusBadRequest,
			&users.HandlerAddParams{},
		},
		{
			"Valid User",
			http.StatusCreated,
			&users.HandlerAddParams{Name: "Name", Email: "email@fake.com", Password: "password"},
		},
		{
			"Duplicate Email",
			http.StatusConflict,
			&users.HandlerAddParams{Name: "Name", Email: "email@fake.com", Password: "password"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			rec := callHandlerAdd(t, tc.params)
			assert.Equal(t, tc.code, rec.Code)

			if rec.Code == http.StatusCreated {
				var u auth.User
				if err := json.NewDecoder(rec.Body).Decode(&u); err != nil {
					t.Fatal(err)
				}

				assert.NotEmpty(t, u.ID)
				assert.Equal(t, tc.params.Email, u.Email)
				testhelpers.SaveModel(t, &u)
			}
		})
	}
}

func callHandlerAdd(t *testing.T, params *users.HandlerAddParams) *httptest.ResponseRecorder {
	ri := &testhelpers.RequestInfo{
		Test:     t,
		Endpoint: users.Endpoints[users.EndpointAdd],
		URI:      "/users/",
		Params:   params,
	}

	return testhelpers.NewRequest(ri)
}