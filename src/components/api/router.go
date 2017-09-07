package api

import (
	"github.com/Nivl/go-rest-tools/dependencies"
	"github.com/Nivl/go-rest-tools/router"
	"github.com/Nivl/go-rest-tools/types/apierror"
	"github.com/gorilla/mux"
	"github.com/melvin-laplanche/ml-api/src/components/about"
	"github.com/melvin-laplanche/ml-api/src/components/sessions"
	"github.com/melvin-laplanche/ml-api/src/components/users"
)

var notFoundEndpoint = &router.Endpoint{
	Handler: func(req router.HTTPRequest, deps *router.Dependencies) error {
		return apierror.NewNotFound()
	},
}

// GetRouter return the api router with all the routes
func GetRouter(deps dependencies.Dependencies) *mux.Router {
	r := mux.NewRouter()
	users.SetRoutes(r, deps)
	sessions.SetRoutes(r, deps)
	about.SetRoutes(r, deps)
	r.NotFoundHandler = router.Handler(notFoundEndpoint, deps)
	return r
}
