package routes

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
)

func handleUser(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return baseSingleResponseHandler(controllers.GetUser(c, r, id))
	case "PATCH":
		return baseSingleResponseHandler(controllers.UpdateUser(c, r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleUsers(c context.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, err := controllers.GetUsers(c, r)
		return baseResponseHandler(val, included, count, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the users.
func UsersHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleUsers(c, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func UserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleUser(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
	}
	return
}
