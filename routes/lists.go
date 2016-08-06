package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
)

func handleMediaList(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetMediaList(c, r, id)
	case "PATCH":
		return controllers.UpdateMediaList(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleMediaLists(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetMediaLists(c, r)
	case "POST":
		return controllers.CreateMediaList(c, w, r)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the agencies.
func MediaListsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handleMediaLists(c, w, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Media Lists handling error", err.Error())
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
			permissions.ReturnError(w, http.StatusInternalServerError, "Media List handling error", err.Error())
			return
		}
	}
}
