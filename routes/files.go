package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/files"
	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/parse"
)

func handleFile(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetFile(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleFiles(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetFiles(c, r)
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
		c.Errorf("files error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Files handling error", err.Error())
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
			c.Errorf("file error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
			return
		}
	}
}

func FileActionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, idOk := vars["id"]
	action, actionOk := vars["action"]
	if idOk && actionOk {
		file, err := files.ReadFile(r, id)
		if err != nil {
			c.Errorf("file error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
			return
		}

		if action == "header" {
			val, err := parse.FileToExcelHeader(r, file)
			if err == nil {
				err = json.NewEncoder(w).Encode(val)
			}
		}

		if err != nil {
			c.Errorf("file error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
			return
		}
	}
}
