package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/news-ai/tabulae/models"
)

func handleUser(c appengine.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetCurrentUser(c)
	}
	return nil, fmt.Errorf("method not implemented")
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	val, err := handleUser(c, r)
	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}
	if err != nil {
		c.Errorf("todo error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
