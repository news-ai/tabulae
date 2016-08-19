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

func handleEmailAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "send":
			return controllers.SendEmail(c, r, id)
		}
	}
	return nil, errors.New("method not implemented")
}

func handleEmail(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetEmail(c, r, id)
	case "PATCH":
		return controllers.UpdateSingleEmail(c, r, id)
	}
	return nil, errors.New("method not implemented")
}

func handleEmails(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetEmails(c, r)
	case "POST":
		return controllers.CreateEmail(c, r)
	case "PATCH":
		return controllers.UpdateBatchEmail(c, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func EmailsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handleEmails(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func EmailHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	id := ps.ByName("id")
	val, err := handleEmail(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
	}
	return
}

func EmailActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	id := ps.ByName("id")
	action := ps.ByName("action")
	val, err := handleEmailAction(c, r, id, action)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
	}
	return
}
