package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/melvin-laplanche/ml-api/src/components/blog"
	"github.com/melvin-laplanche/ml-api/src/components/sessions"
	"github.com/melvin-laplanche/ml-api/src/components/users"
	"github.com/melvin-laplanche/ml-api/src/mlhttp"
)

func notFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := fmt.Sprintf(`{"error":"%s"}`, http.StatusText(http.StatusNotFound))
	mlhttp.ErrorJSON(w, err, http.StatusNotFound)
}

// GetRouter return the api router with all the routes
func GetRouter() *mux.Router {
	baseURI := ""
	r := mux.NewRouter()
	blog.SetRoutes(baseURI, r)
	users.SetRoutes(baseURI, r)
	sessions.SetRoutes(baseURI, r)
	r.NotFoundHandler = http.HandlerFunc(notFound)

	return r
}
