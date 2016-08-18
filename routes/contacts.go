package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
)

func handleContactAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "diff":
			return controllers.GetDiff(c, r, id)
		case "update":
			return controllers.UpdateContactToParent(c, r, id)
		case "sync":
			return controllers.LinkedInSync(c, r, id)
		}
	}
	return nil, errors.New("method not implemented")
}

func handleContact(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetContact(c, r, id)
	case "PATCH":
		return controllers.UpdateSingleContact(c, r, id)
	}
	return nil, errors.New("method not implemented")
}

func handleContacts(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		limit, offset, err := GetPagination(r)
		if err != nil {
			return nil, err
		}
		return controllers.GetContacts(c, r, limit, offset)
	case "POST":
		return controllers.CreateContact(c, r)
	case "PATCH":
		return controllers.UpdateBatchContact(c, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func ContactsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handleContacts(c, w, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		val, err := handleContact(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
			return
		}
	}
}

func ContactActionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, idOk := vars["id"]
	action, actionOk := vars["action"]
	if idOk && actionOk {
		val, err := handleContactAction(c, r, action, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
			return
		}
	}
}
