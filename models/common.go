package models

import (
	"errors"
	"net/http"
	"net/url"

	"appengine"

	"github.com/gorilla/context"
)

func getCurrentUser(c appengine.Context, r *http.Request) (User, error) {
	// Get the current user
	_, ok := context.GetOk(r, "user")
	if !ok {
		return User{}, errors.New("No user logged in")
	}
	user := context.Get(r, "user").(User)
	return user, nil
}

func stripQueryString(inputUrl string) string {
	u, err := url.Parse(inputUrl)
	if err != nil {
		return inputUrl
	}
	u.RawQuery = ""
	return u.String()
}
