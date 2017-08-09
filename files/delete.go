package files

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/cloud/storage"

	"github.com/news-ai/tabulae/models"
)

func DeleteFile(r *http.Request, file models.File) error {
	c := appengine.NewContext(r)

	bucketName := ""
	if file.ListId == 0 {
		bucketName = "tabulae-email-attachment"
	}

	bucket, err := getStorageBucket(r, bucketName)
	if err != nil {
		return err
	}

	client, err := storage.NewClient(c)
	defer client.Close()
	if err != nil {
		return err
	}

	// Setup the bucket to upload the file
	clientBucket := client.Bucket(bucket)
	err = clientBucket.Object(file.FileName).Delete(c)
	if err != nil {
		return err
	}

	return nil
}
