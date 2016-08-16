package parse

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/news-ai/tabulae/models"

	"golang.org/x/net/context"

	"github.com/extrame/xls"
)

func xlsGetCustomFields(r *http.Request, c context.Context, numberOfColumns int, singleRow *xls.Row, headers []string) []string {
	var customFields []string

	for x := 0; x < numberOfColumns; x++ {
		columnName := headers[x]
		if !customOrNative(columnName) {
			customFields = append(customFields, columnName)
		}
	}
	return customFields
}

func xlsRowToContact(r *http.Request, c context.Context, singleRow *xls.Row, headers []string) (models.Contact, error) {
	return models.Contact{}, nil
}

func XlsToContactList(r *http.Request, file []byte, headers []string, mediaListid int64) ([]models.Contact, []string, error) {
	return []models.Contact{}, []string{}, nil
}

func XlsFileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	readerFile := bytes.NewReader(file)
	workbook, err := xls.OpenReader(readerFile, "utf-8")
	if err != nil {
		return []Column{}, err
	}

	sheet := workbook.GetSheet(0)
	if sheet == nil {
		return []Column{}, errors.New("Sheet is empty")
	}

	// Number of rows to consider
	numberOfRows := 15
	if int(sheet.MaxRow) < numberOfRows+1 {
		numberOfRows = int(sheet.MaxRow)
	}

	numberOfColumns := len(sheet.Rows[0].Cols)
	columns := make([]Column, numberOfColumns)

	for i := 0; i <= numberOfRows; i++ {
		row := sheet.Rows[uint16(i)]
		for x := 0; x < numberOfColumns; x++ {
			currentRow := row.Cols[uint16(x)]
			cellName := ""
			if currentRow != nil {
				cellName = row.Cols[uint16(x)].String(workbook)[0]
			}
			columns[x].Rows = append(columns[x].Rows, strings.Trim(cellName, " "))
		}
	}

	return columns, nil
}
