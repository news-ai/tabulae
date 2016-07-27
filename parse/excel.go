package parse

import (
	"net/http"
	"strings"

	"github.com/tealeg/xlsx"
)

type Column struct {
	Rows []string `json:"rows"`
}

func FileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	xlFile, err := xlsx.OpenBinary(file)
	if err != nil {
		return []Column{}, err
	}

	sheet := xlFile.Sheets[0]

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
