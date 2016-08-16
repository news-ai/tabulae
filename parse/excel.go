package parse

import (
	"errors"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

func FileToExcelHeader(r *http.Request, file []byte, contentType string) ([]Column, error) {
	c := appengine.NewContext(r)
	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		return xlsFileToExcelHeader(r, file)
	}
	return xlsxFileToExcelHeader(r, file)
}

func ExcelHeadersToListModel(r *http.Request, file []byte, headers []string, mediaListid int64, contentType string) (models.MediaList, error) {
	c := appengine.NewContext(r)

	contacts := []models.Contact{}
	customFields := []string{}
	err := errors.New("")

	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		contacts, customFields, err = xlsToContactList(r, file, headers, mediaListid)
		if err != nil {
			return models.MediaList{}, err
		}
	} else {
		contacts, customFields, err = xlsxToContactList(r, file, headers, mediaListid)
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
	mediaList, err := controllers.GetMediaList(c, r, mediaListId)
	mediaList.Contacts = contactIds
	mediaList.Fields = headers
	mediaList.CustomFields = customFields
	mediaList.Save(c)
	return mediaList, nil
}
