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

func getStorageBucket(c appengine.Context, bucket string) (string, error) {
	if bucket == "" {
		var err error
		if bucket, err = file.DefaultBucketName(c); err != nil {
			return bucket, err
		}
		return bucket, nil
	}
	return bucket, nil
}
