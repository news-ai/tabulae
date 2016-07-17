package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/models"
)

func handleUser(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetUser(c, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleUsers(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetUsers(c, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	fmt.Println(vars)
	id, ok := vars["id"]
	if ok {
		val, err := handleUser(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			c.Errorf("user error: %#v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		val, err := handleUsers(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			c.Errorf("user error: %#v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
