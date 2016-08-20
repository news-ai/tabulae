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

func handleTemplate(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetTemplate(c, r, id)
	case "PATCH":
		return controllers.UpdateTemplate(c, r, id)
	}
	return nil, errors.New("method not implemented")
}

func handleTemplates(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetTemplates(c, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the agencies.
func TemplatesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleTemplates(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Template handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func TemplateHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleTemplate(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Template handling error", err.Error())
	}
	return
}
