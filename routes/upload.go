package routes

import (
	"net/http"

	"appengine"

	"github.com/news-ai/tabulae/middleware"
)

// Handler for when the user wants all the users.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	bucket, err := getStorageBucket(c, "")
	if err != nil {
		c.Errorf("upload error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
		return
	}

	c.Infof("%v", bucket)

	middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "")
}
