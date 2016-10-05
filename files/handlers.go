package files

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/parse"

	"github.com/news-ai/web/utilities"
)

func HandleMediaListActionUpload(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		return nil, nil, err
	}

	userId := strconv.FormatInt(user.Id, 10)

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	noSpaceFileName := ""
	if handler.Filename != "" {
		noSpaceFileName = strings.Replace(handler.Filename, " ", "", -1)
	}

	fileName := strings.Join([]string{userId, id, utilities.RandToken(), noSpaceFileName}, "-")
	val, err := UploadFile(r, fileName, file, userId, id, handler.Header.Get("Content-Type"))
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return val, nil, nil
}

func HandleFileUploadHeaders(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	var fileOrder models.FileOrder
	err := decoder.Decode(&fileOrder)
	if err != nil {
		return nil, nil, err
	}

	// Get & write file
	file, _, err := controllers.GetFile(c, r, id)
	if err != nil {
		return nil, nil, err
	}

	if file.Imported {
		return nil, nil, err
	}

	file.Order = fileOrder.Order

	// Read file
	byteFile, contentType, err := ReadFile(r, id)
	if err != nil {
		return nil, nil, err
	}

	// Import the file
	_, err = parse.ExcelHeadersToListModel(r, byteFile, file.Order, file.ListId, contentType)
	if err != nil {
		return nil, nil, err
	}

	// Return the file
	file.Imported = true
	val, err := file.Save(c)
	if err != nil {
		return nil, nil, err
	}

	// Return value
	if err == nil {
		return val, nil, nil
	}

	return nil, nil, err
}

func HandleFileGetHeaders(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	file, contentType, err := ReadFile(r, id)
	if err != nil {
		return nil, nil, err
	}

	// Parse file headers and report to API
	val, err := parse.FileToExcelHeader(r, file, contentType)
	if err == nil {
		return val, nil, nil
	}

	return nil, nil, err
}

func HandleFileGetSheets(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	file, contentType, err := ReadFile(r, id)
	if err != nil {
		return nil, nil, err
	}

	// Parse file headers and report to API
	val, err := parse.FileToExcelSheets(r, file, contentType)
	if err == nil {
		return val, nil, nil
	}

	return nil, nil, err
}
