package organizations_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/Nivl/go-rest-tools/router"
	"github.com/Nivl/go-rest-tools/router/guard/testguard"
	"github.com/Nivl/go-rest-tools/router/mockrouter"
	"github.com/Nivl/go-rest-tools/router/params"
	"github.com/Nivl/go-rest-tools/router/testrouter"
	"github.com/Nivl/go-rest-tools/security/auth"
	"github.com/Nivl/go-rest-tools/storage/db/mockdb"
	"github.com/Nivl/go-rest-tools/types/apierror"
	"github.com/Nivl/go-rest-tools/types/ptrs"
	"github.com/melvin-laplanche/ml-api/src/components/about/organizations"
	"github.com/melvin-laplanche/ml-api/src/components/about/organizations/testorganizations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateInvalidParams(t *testing.T) {
	testCases := []testguard.InvalidParamsTestCase{
		{
			Description: "Should fail on missing ID",
			MsgMatch:    params.ErrMsgMissingParameter,
			FieldName:   "id",
			Sources: map[string]url.Values{
				"url": url.Values{},
			},
		},
		{
			Description: "Should fail on invalid ID",
			MsgMatch:    params.ErrMsgInvalidUUID,
			FieldName:   "id",
			Sources: map[string]url.Values{
				"url": url.Values{
					"id": []string{"xxx"},
				},
			},
		},
		{
			Description: "Should fail on not nil but empty name",
			MsgMatch:    params.ErrMsgEmptyParameter,
			FieldName:   "name",
			Sources: map[string]url.Values{
				"url": url.Values{
					"id": []string{"aa44ca86-553e-4e16-8c30-2e50e63f7eaa"},
				},
				"form": url.Values{
					"name": []string{"     "},
				},
			},
		},
		{
			Description: "Should fail on not nil but invalid website",
			MsgMatch:    params.ErrMsgInvalidURL,
			FieldName:   "website",
			Sources: map[string]url.Values{
				"url": url.Values{
					"id": []string{"aa44ca86-553e-4e16-8c30-2e50e63f7eaa"},
				},
				"form": url.Values{
					"name":    []string{"valid name"},
					"website": []string{"not-a-url"},
				},
			},
		},
		{
			Description: "Should fail on not nil but invalid in_trash",
			MsgMatch:    params.ErrMsgInvalidBoolean,
			FieldName:   "in_trash",
			Sources: map[string]url.Values{
				"url": url.Values{
					"id": []string{"aa44ca86-553e-4e16-8c30-2e50e63f7eaa"},
				},
				"form": url.Values{
					"in_trash": []string{"not-a-boolean"},
				},
			},
		},
	}

	g := organizations.Endpoints[organizations.EndpointUpdate].Guard
	testguard.InvalidParams(t, g, testCases)
}

func TestUpdateValidParams(t *testing.T) {
	testCases := []struct {
		description string
		sources     map[string]url.Values
	}{
		{
			"Should work with only a valid uuid",
			map[string]url.Values{
				"url": url.Values{
					"id": []string{"aa44ca86-553e-4e16-8c30-2e50e63f7eaa"},
				},
				"form": url.Values{},
			},
		},
		{
			"Should work with only a valid name",
			map[string]url.Values{
				"url": url.Values{
					"id": []string{"aa44ca86-553e-4e16-8c30-2e50e63f7eaa"},
				},
				"form": url.Values{
					"name": []string{"valid name"},
				},
			},
		},
		{
			"Should work with only a valid in_trash",
			map[string]url.Values{
				"url": url.Values{
					"id": []string{"aa44ca86-553e-4e16-8c30-2e50e63f7eaa"},
				},
				"form": url.Values{
					"in_trash": []string{"0"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			endpts := organizations.Endpoints[organizations.EndpointUpdate]
			_, err := endpts.Guard.ParseParams(tc.sources, nil)
			assert.NoError(t, err)
		})
	}
}

func TestUpdateAccess(t *testing.T) {
	testCases := []testguard.AccessTestCase{
		{
			Description: "Should fail for anonymous users",
			User:        nil,
			ErrCode:     http.StatusUnauthorized,
		},
		{
			Description: "Should fail for logged users",
			User:        &auth.User{ID: "48d0c8b8-d7a3-4855-9d90-29a06ef474b0"},
			ErrCode:     http.StatusForbidden,
		},
		{
			Description: "Should work for admin users",
			User:        &auth.User{ID: "48d0c8b8-d7a3-4855-9d90-29a06ef474b0", IsAdmin: true},
			ErrCode:     0,
		},
	}

	g := organizations.Endpoints[organizations.EndpointUpdate].Guard
	testguard.AccessTest(t, g, testCases)
}

func TestUpdateHappyPath(t *testing.T) {
	handlerParams := &organizations.UpdateParams{
		ID:        "48d0c8b8-d7a3-4855-9d90-29a06ef474b0",
		Name:      ptrs.NewString("Google"),
		ShortName: ptrs.NewString("googl"),
		Website:   ptrs.NewString("https://google.com"),
		InTrash:   ptrs.NewBool(true),
	}

	// Mock the database & add expectations
	mockDB := &mockdb.Connection{}
	mockDB.ExpectGet("*organizations.Organization", func(args mock.Arguments) {
		org := args.Get(0).(*organizations.Organization)
		org.ID = "48d0c8b8-d7a3-4855-9d90-29a06ef474b0"
		org.Name = "Not Google"
	})
	mockDB.ExpectUpdate("*organizations.Organization")

	// Mock the response & add expectations
	res := new(mockrouter.HTTPResponse)
	res.ExpectOk("*organizations.Payload", func(args mock.Arguments) {
		org := args.Get(0).(*organizations.Payload)
		assert.Equal(t, handlerParams.ID, org.ID, "ID should have not changed")
		assert.Equal(t, *handlerParams.Name, org.Name, "name should have been updated")
		assert.Equal(t, *handlerParams.Website, org.Website, "Website should have been updated")
		assert.Equal(t, *handlerParams.ShortName, org.ShortName, "ShortName should have been updated")
		assert.NotNil(t, org.DeletedAt, "DeletedAt should have been set")
	})

	// Mock the request & add expectations
	req := new(mockrouter.HTTPRequest)
	req.On("Response").Return(res)
	req.On("Params").Return(handlerParams)

	// call the handler
	err := organizations.Update(req, &router.Dependencies{DB: mockDB})

	// Assert everything
	assert.Nil(t, err, "the handler should not have fail")
	mockDB.AssertExpectations(t)
	req.AssertExpectations(t)
	res.AssertExpectations(t)
}

func TestUpdateConflictName(t *testing.T) {
	p := &testrouter.ConflictTestParams{
		StructConflicting: "*organizations.Organization",
		FieldConflicting:  "name",
		Handler:           organizations.Update,
		HandlerParams: &organizations.UpdateParams{
			Name: ptrs.NewString("Google"),
		},
	}
	testrouter.ConflictUpdateTest(t, p)
}

func TestUpdateConflictShortName(t *testing.T) {
	p := &testrouter.ConflictTestParams{
		StructConflicting: "*organizations.Organization",
		FieldConflicting:  "short_name",
		Handler:           organizations.Update,
		HandlerParams: &organizations.UpdateParams{
			ShortName: ptrs.NewString("googl"),
		},
	}
	testrouter.ConflictUpdateTest(t, p)
}

func TestUpdateNoDBCon(t *testing.T) {
	handlerParams := &organizations.UpdateParams{
		ID:        "48d0c8b8-d7a3-4855-9d90-29a06ef474b0",
		Name:      ptrs.NewString("Google"),
		ShortName: ptrs.NewString("googl"),
		Website:   ptrs.NewString("https://google.com"),
		InTrash:   ptrs.NewBool(true),
	}

	// Mock the database & add expectations
	mockDB := &mockdb.Connection{}
	mockDB.ExpectGet("*organizations.Organization", func(args mock.Arguments) {
		exp := args.Get(0).(*organizations.Organization)
		*exp = *(testorganizations.New())
	})
	mockDB.ExpectUpdateError("*organizations.Organization")

	// Mock the request & add expectations
	req := new(mockrouter.HTTPRequest)
	req.On("Params").Return(handlerParams)

	// call the handler
	err := organizations.Update(req, &router.Dependencies{DB: mockDB})

	// Assert everything
	assert.Error(t, err, "the handler should have fail")
	mockDB.AssertExpectations(t)
	req.AssertExpectations(t)

	apiError := apierror.Convert(err)
	assert.Equal(t, http.StatusInternalServerError, apiError.HTTPStatus())
}

func TestUpdateUnexisting(t *testing.T) {
	handlerParams := &organizations.UpdateParams{
		ID:        "48d0c8b8-d7a3-4855-9d90-29a06ef474b0",
		Name:      ptrs.NewString("Google"),
		ShortName: ptrs.NewString("googl"),
		Website:   ptrs.NewString("https://google.com"),
		InTrash:   ptrs.NewBool(true),
	}

	// Mock the database & add expectations
	mockDB := &mockdb.Connection{}
	mockDB.ExpectGetNotFound("*organizations.Organization")

	// Mock the request & add expectations
	req := new(mockrouter.HTTPRequest)
	req.On("Params").Return(handlerParams)

	// call the handler
	err := organizations.Update(req, &router.Dependencies{DB: mockDB})

	// Assert everything
	assert.Error(t, err, "the handler should have fail")
	mockDB.AssertExpectations(t)
	req.AssertExpectations(t)

	apiError := apierror.Convert(err)
	assert.Equal(t, http.StatusNotFound, apiError.HTTPStatus())
}
