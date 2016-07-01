package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

// SearchPerson handles a search request for a person.
func SearchPerson(ctx context.Context, env *Env, res http.ResponseWriter, req *http.Request) error {
	query := req.URL.Query().Get("query")

	env.Debugf("listing people that match %v", query)

	people, err := env.DB.FuzzyFindPersons(query)
	if err != nil {
		return err
	}

	return httpWriteJSON(res, http.StatusOK, people)
}

// SearchHandler adds routes to the for ghenga API in the given environment to r.
func SearchHandler(ctx context.Context, env *Env, r *mux.Router) {
	r.Handle("/api/search/person", Handle(ctx, env, RequireAuth(SearchPerson))).Methods("GET")
}
