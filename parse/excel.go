package parse

import (
	"errors"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"

	"github.com/news-ai/goexcel"
)

func FileToExcelHeader(r *http.Request, file []byte, contentType string) ([]goexcel.Column, error) {
	c := appengine.NewContext(r)
	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		return goexcel.XlsFileToExcelHeader(r, file)
	}
	return goexcel.XlsxFileToExcelHeader(r, file)
}

func ExcelHeadersToListModel(r *http.Request, file []byte, headers []string, mediaListid int64, contentType string) (models.MediaList, error) {
	c := appengine.NewContext(r)

	contacts := []models.Contact{}
	var customFields map[string]bool
	err := errors.New("")

	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		contacts, customFields, err = goexcel.XlsToContactList(r, file, headers, mediaListid)
		if err != nil {
			return models.MediaList{}, err
		}
	} else {
		contacts, customFields, err = goexcel.XlsxToContactList(r, file, headers, mediaListid)
		if err != nil {
			return models.MediaList{}, err
		}
	}

	// Batch create all the contacts
	contactIds, err := controllers.BatchCreateContactsForExcelUpload(c, r, contacts)
	if err != nil {
		return models.MediaList{}, err
	}

	mediaListId := utils.IntIdToString(mediaListid)
	mediaList, _, err := controllers.GetMediaList(c, r, mediaListId)
	mediaList.Contacts = contactIds
	customFieldsList := []models.CustomFieldsMap{}
	for i := 0; i < len(headers); i++ {
		customField := models.CustomFieldsMap{}
		customField.Name = headers[i]
		customField.Value = headers[i]
		if _, ok := customFields[headers[i]]; ok {
			customField.CustomField = true
		} else {
			customField.CustomField = false
		}
		customField.Hidden = false
		customFieldsList = append(customFieldsList, customField)
	}

	mediaList.FieldsMap = customFieldsList

	mediaList.Save(c)
	return mediaList, nil
}
