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

func UploadFile(r *http.Request, fileName string, file io.Reader, userId, listId, contentType string) (models.File, error) {
	c := appengine.NewContext(r)

	bucket, err := getStorageBucket(r, "")
	if err != nil {
		return models.File{}, err
	}

	client, err := storage.NewClient(c)
	defer client.Close()
	if err != nil {
		return models.File{}, err
	}

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

	val, err := controllers.CreateFile(r, fileName, listId, userId)
	if err != nil {
		return models.File{}, err
	}
	return val, nil
}

func UploadImage(r *http.Request, originalFilename string, fileName string, file io.Reader, userId, contentType string) (models.File, error) {
	c := appengine.NewContext(r)

	bucket, err := getImageStorageBucket(r, "tabulae-email-images")
	if err != nil {
		return models.File{}, err
	}

	client, err := storage.NewClient(c)
	defer client.Close()
	if err != nil {
		return models.File{}, err
	}

	// Setup the bucket to upload the file
	clientBucket := client.Bucket(bucket)
	wc := clientBucket.Object(fileName).NewWriter(c)
	wc.ContentType = contentType
	wc.Metadata = map[string]string{
		"x-goog-meta-userid": userId,
	}
	wc.ACL = []storage.ACLRule{{Entity: storage.ACLEntity("project-owners-newsai-1166"), Role: storage.RoleOwner}}
	wc.ACL = append(wc.ACL, storage.ACLRule{Entity: storage.AllUsers, Role: storage.RoleReader})

	wc.CacheControl = "public, max-age=86400"
	wc.ContentDisposition = "inline"

	// Upload the file
	data, err := ioutil.ReadAll(file)
	if _, err := wc.Write(data); err != nil {
		return models.File{}, err
	}
	if err := wc.Close(); err != nil {
		return models.File{}, err
	}

	val, err := controllers.CreateImageFile(r, originalFilename, fileName, userId, bucket)
	if err != nil {
		return models.File{}, err
	}

	return val, nil
}

func UploadAttachment(r *http.Request, originalFilename, fileName string, file io.Reader, userId, emailId, contentType string) (models.File, error) {
	c := appengine.NewContext(r)

	bucket, err := getImageStorageBucket(r, "tabulae-email-attachment")
	if err != nil {
		return models.File{}, err
	}

	client, err := storage.NewClient(c)
	defer client.Close()
	if err != nil {
		return models.File{}, err
	}

	// Setup the bucket to upload the file
	clientBucket := client.Bucket(bucket)
	wc := clientBucket.Object(fileName).NewWriter(c)
	wc.ContentType = contentType
	wc.Metadata = map[string]string{
		"x-goog-meta-userid":  userId,
		"x-goog-meta-emailId": emailId,
	}
	wc.ACL = []storage.ACLRule{{Entity: storage.ACLEntity("project-owners-newsai-1166"), Role: storage.RoleOwner}}
	wc.CacheControl = "public, max-age=86400"

	// Upload the file
	data, err := ioutil.ReadAll(file)
	if _, err := wc.Write(data); err != nil {
		return models.File{}, err
	}
	if err := wc.Close(); err != nil {
		return models.File{}, err
	}

	val, err := controllers.CreateAttachmentFile(r, originalFilename, fileName, emailId, userId)
	if err != nil {
		return models.File{}, err
	}

	return val, nil
}
