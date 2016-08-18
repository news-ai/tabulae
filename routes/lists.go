package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"
	"github.com/news-ai/tabulae/permissions"
)

var (
	errMediaListHandling = "Media List handling error"
)

func handleMediaListActions(c context.Context, r *http.Request, id string, action string, limit int, offset int) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "contacts":
			return controllers.GetContactsForList(c, r, id, limit, offset)
		}
	case "POST":
		switch action {
		case "upload":
			user, err := GetUser(r)
			if err != nil {
				return nil, err
			}
			return files.HandleMediaListActionUpload(c, r, id, user, limit, offset)
		}
	}
	return nil, errors.New("method not implemented")
}

func handleMediaList(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetMediaList(c, r, id)
	case "PATCH":
		return controllers.UpdateMediaList(c, r, id)
	}
	return nil, errors.New("method not implemented")
}

func handleMediaLists(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetMediaLists(c, r)
	case "POST":
		return controllers.CreateMediaList(c, w, r)
	}
	return nil, errors.New("method not implemented")
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
		permissions.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
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
			permissions.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
			return
		}
	}
}

// Handler for when the user wants to perform an action on the lists
func MediaListActionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// Get Id and Action
	vars := mux.Vars(r)
	id, idOk := vars["id"]
	action, actionOk := vars["action"]

	if idOk && actionOk {
		limit, offset, err := GetPagination(r)
		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
			return
		}

		val, err := handleMediaListActions(c, r, id, action, limit, offset)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
			return
		}
	}
}
