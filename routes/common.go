package routes

import (
	"errors"
	"net/http"

	"appengine"
	"appengine/file"
)

func IsAdmin(w http.ResponseWriter, r *http.Request) error {
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
