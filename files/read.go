package files

import (
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/cloud/storage"
)

func ReadFile(r *http.Request, fileId string) ([]byte, string, error) {
	c := appengine.NewContext(r)

	bucket, err := getStorageBucket(r, "")
	if err != nil {
		return nil, "", err
	}

	client, err := storage.NewClient(c)
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	file, err := getFile(r, fileId)
	if err != nil {
		return nil, "", err
	}

	clientBucket := client.Bucket(bucket)
	rc, err := clientBucket.Object(file.FileName).NewReader(c)
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, "", err
	}

	return data, rc.ContentType(), nil
}
