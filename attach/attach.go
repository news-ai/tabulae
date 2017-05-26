package attach

import (
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/models"
)

func ReadAttachment(r *http.Request, file models.File) ([]byte, string, string, error) {
	c := appengine.NewContext(r)

	client, err := storage.NewClient(c)
	if err != nil {
		return nil, "", "", err
	}
	defer client.Close()

	clientBucket := client.Bucket("tabulae-email-attachment")
	rc, err := clientBucket.Object(file.FileName).NewReader(c)
	if err != nil {
		return nil, "", "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, "", "", err
	}

	return data, rc.ContentType(), file.OriginalName, nil
}

func GetAttachmentsForEmail(r *http.Request, email models.Email, files []models.File) ([][]byte, []string, []string, error) {
	if len(files) == 0 {
		return [][]byte{}, []string{}, []string{}, nil
	}

	c := appengine.NewContext(r)
	bytesArray := [][]byte{}
	attachmentTypes := []string{}
	fileNames := []string{}
	for i := 0; i < len(files); i++ {
		currentBytes, attachmentType, fileName, err := ReadAttachment(r, files[i])
		if err == nil {
			bytesArray = append(bytesArray, currentBytes)
			attachmentTypes = append(attachmentTypes, attachmentType)
			fileNames = append(fileNames, fileName)
		} else {
			log.Errorf(c, "%v", err)
		}
	}

	return bytesArray, attachmentTypes, fileNames, nil
}
