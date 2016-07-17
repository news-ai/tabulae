package routes

import (
	"errors"
	"net/http"

	"appengine"
	"appengine/user"

	"github.com/news-ai/tabulae/middleware"
)

func GetUser(c appengine.Context, w http.ResponseWriter) *user.User {
	u := user.Current(c)
	return u
}

func IsAdmin(w http.ResponseWriter, r *http.Request, u *user.User) error {
	if !u.Admin {
		middleware.ReturnError(w, http.StatusUnauthorized, "Permission denied", "Admin login only")
		return errors.New("Admin login only")
	}
	return nil
}
