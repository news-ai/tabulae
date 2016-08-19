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
	"github.com/news-ai/tabulae/utils"
)

func HandleMediaListActionUpload(c context.Context, r *http.Request, id string, user models.User) (interface{}, error) {
	userId := strconv.FormatInt(user.Id, 10)

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	noSpaceFileName := ""
	if handler.Filename != "" {
		noSpaceFileName = strings.Replace(handler.Filename, " ", "", -1)
	}

	fileName := strings.Join([]string{userId, id, utils.RandToken(), noSpaceFileName}, "-")
	val, err := UploadFile(r, fileName, file, userId, id, handler.Header.Get("Content-Type"))
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	return val, nil
}

func HandleFileUploadHeaders(c context.Context, r *http.Request, id string) (interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	var fileOrder models.FileOrder
	err := decoder.Decode(&fileOrder)
	if err != nil {
		return nil, err
	}

	// Get & write file
	file, err := controllers.GetFile(c, r, id)
	if err != nil {
		return nil, err
	}

	if file.Imported {
		return nil, err
	}

	file.Order = fileOrder.Order

	// Read file
	byteFile, contentType, err := ReadFile(r, id)
	if err != nil {
		return nil, err
	}

	// Import the file
	_, err = parse.ExcelHeadersToListModel(r, byteFile, file.Order, file.ListId, contentType)
	if err != nil {
		return nil, err
	}

	// Return the file
	file.Imported = true
	val, err := file.Save(c)
	if err != nil {
		return nil, err
	}

	// Return value
	if err == nil {
		return val, nil
	}

	return nil, err
}

func HandleFileGetHeaders(c context.Context, r *http.Request, id string) (interface{}, error) {
	file, contentType, err := ReadFile(r, id)
	if err != nil {
		return nil, err
	}

	// Parse file headers and report to API
	val, err := parse.FileToExcelHeader(r, file, contentType)
	if err == nil {
		return val, nil
	}

	return nil, err
}
