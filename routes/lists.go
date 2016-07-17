package routes

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"appengine"

// 	"github.com/news-ai/tabulae/models"
// )

// func handleLists(c appengine.Context, r *http.Request) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		return models.GetUserLists(c)
// 	}
// 	return nil, fmt.Errorf("method not implemented")
// }

// func UserHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	val, err := handleLists(c, r)
// 	if err == nil {
// 		err = json.NewEncoder(w).Encode(val)
// 	}
// 	if err != nil {
// 		c.Errorf("user error: %#v", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
