package parse

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/goexcel"
	"github.com/news-ai/web/utilities"
)

func FileToExcelHeader(r *http.Request, file []byte, contentType string) ([]goexcel.Column, error) {
	c := appengine.NewContext(r)
	return goexcel.FileToExcelHeader(c, r, file, contentType)
}

func ExcelHeadersToListModel(r *http.Request, file []byte, headers []string, mediaListid int64, contentType string) (models.MediaList, error) {
	c := appengine.NewContext(r)

	// Batch create all the contacts
	contacts, customFields, err := goexcel.HeadersToListModel(c, r, file, headers, contentType)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	contactIds, err := controllers.BatchCreateContactsForExcelUpload(c, r, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	mediaListId := utilities.IntIdToString(mediaListid)
	mediaList, _, err := controllers.GetMediaList(c, r, mediaListId)
	mediaList.Contacts = contactIds
	for i := 0; i < len(headers); i++ {
		if _, ok := customFields[headers[i]]; ok {
			customField := models.CustomFieldsMap{}
			customField.Name = headers[i]
			customField.Value = headers[i]
			customField.CustomField = true
			customField.Hidden = false
			mediaList.FieldsMap = append(mediaList.FieldsMap, customField)
		}
	}

	mediaList.Save(c)
	return mediaList, nil
}
