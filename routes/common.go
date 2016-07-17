package routes

import (
	"errors"
	"net/http"

	"appengine"
	"appengine/user"
)

func GetUser(c appengine.Context, w http.ResponseWriter) *user.User {
	u := user.Current(c)
	return u
}

func ListAllowed(w http.ResponseWriter, r *http.Request, u *user.User) error {
	if !u.Admin {
		http.Error(w, "Admin login only", http.StatusUnauthorized)
		return errors.New("Admin login only")
	}
	return nil
}
