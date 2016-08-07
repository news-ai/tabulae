package routes

import (
	"errors"
	"net/http"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
)

func IsAdmin(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		return errors.New("Admin login only")
	}
	if user.IsAdmin {
		return nil
	}
	return errors.New("Admin login only")
}

func GetUser(r *http.Request) (models.User, error) {
	c := appengine.NewContext(r)
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		return models.User{}, errors.New("Admin login only")
	}
	return user, nil
}
