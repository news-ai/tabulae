package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/models"
)

func handleAgency(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetAgency(c, id)
		// case "PATCH":
		// 	return models.UpdateAgency(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleAgencies(c appengine.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetAgencies(c)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the agencies.
func AgenciesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	u := GetUser(c, w)

	err := IsAdmin(w, r, u)
	if err != nil {
		middleware.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
		return
	}

	val, err := handleAgencies(c, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		c.Errorf("agency error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Agency handling error", err.Error())
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
			c.Errorf("agency error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Agency handling error", err.Error())
			return
		}
	}
}
