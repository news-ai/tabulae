package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/parse"
	"github.com/news-ai/tabulae/permissions"
)

func handleFile(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetFile(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleFiles(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetFiles(c, r)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the files.
func FilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handleFiles(c, w, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Files handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /files/<id> route.
func FileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		val, err := handleFile(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
			return
		}
	}
}

func FileActionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// If there is an ID
	vars := mux.Vars(r)
	id, idOk := vars["id"]
	action, actionOk := vars["action"]
	if idOk && actionOk {
		if action == "headers" {
			switch r.Method {
			case "GET":
				// Read file
				file, contentType, err := files.ReadFile(r, id)
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}

				// Parse file headers and report to API
				val, err := parse.FileToExcelHeader(r, file)
				if err == nil {
					err = json.NewEncoder(w).Encode(val)
					return
				}
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}
			case "POST":
				// De-serialize fileOrder from POST data
				c := appengine.NewContext(r)
				decoder := json.NewDecoder(r.Body)
				var fileOrder models.FileOrder
				err := decoder.Decode(&fileOrder)
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}

				// Get & write file
				file, err := controllers.GetFile(c, r, id)
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}

				if file.Imported {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", "File has already been imported")
					return
				}

				file.Order = fileOrder.Order

				// Read file
				byteFile, err := files.ReadFile(r, id)
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}

				// Import the file
				_, err = parse.ExcelHeadersToListModel(r, byteFile, file.Order, file.ListId)
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}

				// Return the file
				file.Imported = true
				val, err := file.Save(c)
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
					return
				}

				// Return value
				if err == nil {
					err = json.NewEncoder(w).Encode(val)
					return
				}
			default:
				permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", "method not implemented")
				return
			}
		}

		permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", "method not implemented")
		return
	}
}
