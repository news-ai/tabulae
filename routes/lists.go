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

func handleMediaList(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetMediaList(c, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleMediaLists(c appengine.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetMediaLists(c)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the agencies.
func MediaListsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	u := GetUser(c, w)

	err := IsAdmin(w, r, u)
	if err != nil {
		return
	}

	val, err := handleMediaLists(c, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		c.Errorf("media lists error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Media Lists handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func MediaListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		val, err := handleMediaList(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			c.Errorf("media list error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Media List handling error", err.Error())
			return
		}
	}
}
