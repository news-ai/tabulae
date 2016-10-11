package files

import (
	"io"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/cloud/storage"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
)

func DeleteFile(r *http.Request, fileName string) errore {
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
	wc := clientBucket.Object(fileName).Delete(c)
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}
