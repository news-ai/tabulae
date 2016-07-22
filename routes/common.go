package routes

import (
	"errors"
	"net/http"

	"appengine"
	"appengine/file"
	"appengine/user"
)

func GetUser(c appengine.Context, w http.ResponseWriter) *user.User {
	u := user.Current(c)
	return u
}

func IsAdmin(w http.ResponseWriter, r *http.Request, u *user.User) error {
	if !u.Admin {
		return errors.New("Admin login only")
	}
	return nil
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
