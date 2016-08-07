package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
)

func handleAgency(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetAgency(c, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleAgencies(c context.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetAgencies(c)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the agencies.
func AgenciesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	err := IsAdmin(w, r)
	if err != nil {
		permissions.ReturnError(w, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	val, err := handleAgencies(c, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Agency handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func AgencyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		val, err := handleAgency(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Agency handling error", err.Error())
			return
		}
	}
}
