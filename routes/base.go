package routes

import (
	"net/http"
)

// Handler for when there is a key present after /users/<id> route.
func BaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Check if the user is an administrator
	err := IsAdmin(w, r)
	if err != nil {
		return
	}
}
