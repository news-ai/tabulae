package routes

import (
	"github.com/news-ai/tabulae"
    "github.com/news-ai/tabulae/models"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func handleUser(c appengine.Context, r *http.Request) (interface{}, error) {
    switch r.Method {
    case "GET":
        return models.getCurrentUser(c)
    return nil, fmt.Errorf("method not implemented")
}

func handler(w http.ResponseWriter, r *http.Request) {
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
