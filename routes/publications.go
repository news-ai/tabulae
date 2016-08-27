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

func handlePublication(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return baseSingleResponseHandler(controllers.GetPublication(c, id))
	}
	return nil, errors.New("method not implemented")
}

func handlePublications(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		if len(r.URL.Query()) > 0 {
			if val, ok := r.URL.Query()["name"]; ok && len(val) > 0 {
				return baseSingleResponseHandler(controllers.FilterPublicationByName(c, val[0]))
			}
		}
		val, included, count, err := controllers.GetPublications(c, r)
		return baseResponseHandler(val, included, count, err, r)
	case "POST":
		val, included, count, err := controllers.CreatePublication(c, w, r)
		if count == 1 {
			return baseSingleResponseHandler(val, included, err)
		}
		return baseResponseHandler(val, included, count, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the agencies.
func PublicationsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handlePublications(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func PublicationHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handlePublication(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
	}
	return
}
