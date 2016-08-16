package routes

import (
	"errors"
	"net/http"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

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

func GetPagination(r *http.Request) (int, int, error) {
	c := appengine.NewContext(r)

	limit := 20
	offset := 0

	queryLimit := r.URL.Query().Get("limit")
	queryOffset := r.URL.Query().Get("offset")
	err := errors.New("")

	// check if query exists
	if len(queryLimit) != 0 {
		limit, err = strconv.Atoi(queryLimit)
		if err != nil {
			log.Errorf(c, "%v", err)
			return limit, offset, err
		}
	}

	// check if offset exists
	if len(queryOffset) != 0 {
		offset, err = strconv.Atoi(queryOffset)
		if err != nil {
			log.Errorf(c, "%v", err)
			return limit, offset, err
		}
	}

	// Boundary checks
	max_limit := 50
	if limit > max_limit {
		limit = max_limit
	}

	return limit, offset, nil
}
