package upload

import (
	"net/http"

	"appengine"
	"appengine/file"
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
