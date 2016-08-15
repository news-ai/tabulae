package parse

import (
	"errors"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"

	"github.com/tealeg/xlsx"
)

var nonCustomHeaders = map[string]bool{
	"firstname":     true,
	"lastname":      true,
	"email":         true,
	"employers":     true,
	"pastemployers": true,
	"notes":         true,
	"linkedin":      true,
	"twitter":       true,
	"instagram":     true,
	"website":       true,
	"blog":          true,
}

type Column struct {
	Rows []string `json:"rows"`
}

func FileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	c := appengine.NewContext(r)

	xlFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	if len(xlFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	sheet := xlFile.Sheets[0]

	if len(sheet.Rows) == 0 {
		err = errors.New("No rows in sheet")
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	// Number of rows to consider
	numberOfRows := 15
	if len(sheet.Rows) < numberOfRows+1 {
		numberOfRows = len(sheet.Rows)
	}

	numberOfColumns := len(sheet.Rows[0].Cells)
	columns := make([]Column, numberOfColumns)

	for _, row := range sheet.Rows[0:numberOfRows] {
		for currentColumn, cell := range row.Cells {
			cellName, _ := cell.String()
			columns[currentColumn].Rows = append(columns[currentColumn].Rows, strings.Trim(cellName, " "))
		}
	}

	return columns, nil
}

func customOrNative(columnName string) bool {
	if _, ok := nonCustomHeaders[columnName]; ok {
		return true
	}
	return false
}

func getCustomFields(r *http.Request, c context.Context, singleRow *xlsx.Row, headers []string) []string {
	var customFields []string

	for columnIndex, _ := range singleRow.Cells {
		columnName := headers[columnIndex]
		if !customOrNative(columnName) {
			customFields = append(customFields, columnName)
		}
	}
	return customFields
}

func rowToContact(r *http.Request, c context.Context, singleRow *xlsx.Row, headers []string) (models.Contact, error) {
	var contact models.Contact

	var employers []int64
	var pastEmployers []int64
	var customFields []models.CustomContactField

	for columnIndex, cell := range singleRow.Cells {
		columnName := headers[columnIndex]
		cellName, _ := cell.String()
		if customOrNative(columnName) {
			switch columnName {
			case "firstname":
				contact.FirstName = cellName
			case "lastname":
				contact.LastName = cellName
			case "email":
				contact.Email = cellName
			case "notes":
				contact.Notes = cellName
			case "employers":
				singleEmployer, err := controllers.FindOrCreatePublication(c, r, cellName)
				if err != nil {
					log.Errorf(c, "employers error: %v", cellName, err)
				}
				employers = append(employers, singleEmployer.Id)
			case "pastemployers":
				singleEmployer, err := controllers.FindOrCreatePublication(c, r, cellName)
				if err != nil {
					log.Errorf(c, "past employers error: %v", cellName, err)
				}
				pastEmployers = append(pastEmployers, singleEmployer.Id)
			case "linkedin":
				contact.LinkedIn = cellName
			case "twitter":
				contact.Twitter = cellName
			case "instagram":
				contact.Instagram = cellName
			case "website":
				contact.Website = cellName
			case "blog":
				contact.Blog = cellName
			}
		} else {
			var customField models.CustomContactField
			customField.Name = columnName
			customField.Value = cellName
			customFields = append(customFields, customField)
		}
	}

	contact.CustomFields = customFields
	contact.Employers = employers
	contact.PastEmployers = pastEmployers
	return contact, nil
}

func ExcelHeadersToListModel(r *http.Request, file []byte, headers []string, mediaListid int64) (models.MediaList, error) {
	c := appengine.NewContext(r)

	xlFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	if len(xlFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	sheet := xlFile.Sheets[0]

	if len(sheet.Rows) == 0 {
		err = errors.New("No rows in sheet")
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	// Number of columns in sheet to compare
	numberOfColumns := len(sheet.Rows[0].Cells)
	if numberOfColumns != len(headers) {
		return models.MediaList{}, errors.New("Number of headers does not match the ones for the sheet")
	}

	mediaListId := utils.IntIdToString(mediaListid)
	mediaList, err := controllers.GetMediaList(c, r, mediaListId)

	// Loop through all the rows
	// Extract information
	contacts := []models.Contact{}
	for _, row := range sheet.Rows {
		contact, err := rowToContact(r, c, row, headers)
		if err != nil {
			return models.MediaList{}, err
		}
		contacts = append(contacts, contact)
	}

	// Batch create all the contacts
	contactIds := []int64{}

	mediaList.Contacts = contactIds
	mediaList.CustomFields = getCustomFields(r, c, sheet.Rows[0], headers)
	mediaList.Save(c)
	return mediaList, nil
}
