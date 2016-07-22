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

func handlePublication(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetPublication(c, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handlePublications(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		if len(r.URL.Query()) > 0 {
			if val, ok := r.URL.Query()["name"]; ok && len(val) > 0 {
				return models.FilterPublicationByName(c, val[0])
			}
		} else {
			u := GetUser(c, w)
			err := IsAdmin(w, r, u)
			if err != nil {
				return nil, err
			}
			return models.GetPublications(c)
		}
	case "POST":
		return models.CreatePublication(c, w, r)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the agencies.
func PublicationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handlePublications(c, w, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		c.Errorf("publication error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func PublicationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		// Convert ID to int64
		val, err := handlePublication(c, r, id)
		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		// If any error from handlePublication function
		if err != nil {
			c.Errorf("publication error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
			return
		}
	}
}
