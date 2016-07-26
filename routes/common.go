package routes

import (
	"errors"
	"net/http"

	"appengine"
	"appengine/file"

	"github.com/news-ai/tabulae/models"
)

func IsAdmin(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	user, err := models.GetCurrentUser(c, r)
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
	user, err := models.GetCurrentUser(c, r)
	if err != nil {
		return models.User{}, errors.New("Admin login only")
	}
	return user, nil
}

func getStorageBucket(r *http.Request, bucket string) (string, error) {
	c := appengine.NewContext(r)
	// In development mode this returns the non-production URL
	if appengine.IsDevAppServer() {
		return "staging.newsai-1166.appspot.com", nil
	}
	if bucket == "" {
		var err error
		if bucket, err = file.DefaultBucketName(c); err != nil {
			return bucket, err
		}
		return bucket, nil
	}
	return bucket, nil
}
