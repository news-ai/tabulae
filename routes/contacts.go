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

func handleContact(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetContact(c, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleContacts(c appengine.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetContacts(c)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the contacts.
func ContactsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	u := GetUser(c, w)

	err := IsAdmin(w, r, u)
	if err != nil {
		return
	}

	val, err := handleContacts(c, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		c.Errorf("contact error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
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
			c.Errorf("contact error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
			return
		}
	}
}
