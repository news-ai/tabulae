package files

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/cloud/storage"
)

func DeleteFile(r *http.Request, fileName string) error {
	c := appengine.NewContext(r)

	bucket, err := getStorageBucket(r, "")
	if err != nil {
		return err
	}

	client, err := storage.NewClient(c)
	if err != nil {
		return err
	}
	defer client.Close()

	// Setup the bucket to upload the file
	clientBucket := client.Bucket(bucket)
	err = clientBucket.Object(fileName).Delete(c)
	if err != nil {
		return err
	}

	return nil
}
