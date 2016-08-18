package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"
	"github.com/news-ai/tabulae/permissions"
)

func handleFileAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "headers":
			return files.HandleFileGetHeaders(c, r, id)
		}
	case "POST":
		switch action {
		case "headers":
			return files.HandleFileUploadHeaders(c, r, id)
		}
	}
	return nil, errors.New("method not implemented")
}

func handleFile(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetFile(c, r, id)
	}
	return nil, errors.New("method not implemented")
}

func handleFiles(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetFiles(c, r)
	}
	return nil, errors.New("method not implemented")
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
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, idOk := vars["id"]
	action, actionOk := vars["action"]
	if idOk && actionOk {
		val, err := handleFileAction(c, r, id, action)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
			return
		}
	}
}
