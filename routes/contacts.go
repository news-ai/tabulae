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

func handleContactAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "diff":
			return baseSingleResponseHandler(controllers.GetDiff(c, r, id))
		case "update":
			return baseSingleResponseHandler(controllers.UpdateContactToParent(c, r, id))
		case "sync":
			return baseSingleResponseHandler(controllers.SocialSync(c, r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleContact(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return baseSingleResponseHandler(controllers.GetContact(c, r, id))
	case "PATCH":
		return baseSingleResponseHandler(controllers.UpdateSingleContact(c, r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleContacts(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, err := controllers.GetContacts(c, r)
		return baseResponseHandler(val, included, count, err, r)
	case "POST":
		val, included, count, err := controllers.CreateContact(c, r)
		return baseResponseHandler(val, included, count, err, r)
	case "PATCH":
		val, included, count, err := controllers.UpdateBatchContact(c, r)
		return baseResponseHandler(val, included, count, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func ContactsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleContacts(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func ContactHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleContact(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
	}
	return
}

func ContactActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	action := ps.ByName("action")
	val, err := handleContactAction(c, r, id, action)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
	}
	return
}
