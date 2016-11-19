package files

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/file"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
)

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

func getImageStorageBucket(r *http.Request, bucket string) (string, error) {
	// In development mode this returns the non-production URL
	if appengine.IsDevAppServer() {
		return "staging-image.newsai-1166.appspot.com", nil
	}
	return bucket, nil
}

func getFile(r *http.Request, fileId string) (models.File, error) {
	c := appengine.NewContext(r)
	file, _, err := controllers.GetFile(c, r, fileId)
	if err != nil {
		return models.File{}, err
	}
	return file, nil
}
