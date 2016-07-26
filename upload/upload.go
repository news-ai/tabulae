package upload

import (
	"io"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/models"

	"google.golang.org/cloud/storage"
)

func UploadFile(r *http.Request, fileName string, file io.Reader, userId string, listId string, contentType string) (models.File, error) {
	c := appengine.NewContext(r)

	bucket, err := getStorageBucket(r, "")
	if err != nil {
		return models.File{}, err
	}

	client, err := storage.NewClient(c)
	if err != nil {
		return models.File{}, err
	}
	defer client.Close()

	// Setup the bucket to upload the file
	clientBucket := client.Bucket(bucket)
	wc := clientBucket.Object(fileName).NewWriter(c)
	wc.ContentType = contentType
	wc.Metadata = map[string]string{
		"x-goog-meta-userid": userId,
		"x-goog-meta-listid": listId,
	}
	wc.ACL = []storage.ACLRule{{Entity: storage.ACLEntity("project-owners-newsai-1166"), Role: storage.RoleOwner}}

	// Upload the file
	data, err := ioutil.ReadAll(file)
	if _, err := wc.Write(data); err != nil {
		return models.File{}, err
	}
	if err := wc.Close(); err != nil {
		return models.File{}, err
	}

	val, err := models.CreateFile(r, fileName, listId, userId)
	if err != nil {
		return models.File{}, err
	}
	return val, nil
}
