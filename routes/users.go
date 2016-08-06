package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/permissions"
)

func handleUser(c appengine.Context, r *http.Request, id string) (models.User, error) {
	switch r.Method {
	case "GET":
		return controllers.GetUser(c, r, id)
	case "PATCH":
		return controllers.UpdateUser(c, r, id)
	}
	return models.User{}, fmt.Errorf("method not implemented")
}

func handleUsers(c appengine.Context, r *http.Request) ([]models.User, error) {
	switch r.Method {
	case "GET":
		return controllers.GetUsers(c)
	}
	return []models.User{}, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the users.
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	err := IsAdmin(w, r)
	if err != nil {
		permissions.ReturnError(w, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	val, err := handleUsers(c, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func UserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		// If the user is trying to get something that is not just their
		// own profile then require them to be an administrator.
		if id != "me" {
			err := IsAdmin(w, r)
			if err != nil {
				permissions.ReturnError(w, http.StatusForbidden, "Forbidden", err.Error())
				return
			}
		}
		val, err := handleUser(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
			return
		}
	}
}
