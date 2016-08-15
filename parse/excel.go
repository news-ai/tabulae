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

var nonCustomHeaders = [...]string{"firstname", "lastname", "email", "employers", "pastemployers", "notes", "linkedin", "twitter", "instagram", "website", "blog"}

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

func rowToContact(r *http.Request, c context.Context, singleRow *xlsx.Row, headers []string) (models.Contact, error) {
	var contact models.Contact

	for currentColumn, cell := range row.Cells {
	}

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
	contacts := []int64{}
	for _, row := range sheet.Rows {
		contact, err := rowToContact(r, c, row, headers)
		if err != nil {
			return models.MediaList{}, err
		}
		contacts = append(contacts, contact.Id)
	}

	mediaList.Contacts = contacts
	mediaList.Save(c)

	return mediaList, nil
}
