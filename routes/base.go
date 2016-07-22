package routes

import (
	"net/http"

	"appengine"
)

// Handler for when there is a key present after /users/<id> route.
func BaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	u := GetUser(c, w)

	// Check if the user is an administrator
	err := IsAdmin(w, r, u)
	if err != nil {
		return
	}
}
