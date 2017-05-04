package routes

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

func handleFileAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "headers":
			return api.BaseSingleResponseHandler(files.HandleFileGetHeaders(c, r, id))
		case "sheets":
			return api.BaseSingleResponseHandler(files.HandleFileGetSheets(c, r, id))
		}
	case "POST":
		switch action {
		case "headers":
			return api.BaseSingleResponseHandler(files.HandleFileUploadHeaders(c, r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleFile(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetFile(c, r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleFiles(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetFiles(c, r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the files.
func FilesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleFiles(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Files handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /files/<id> route.
func FileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleFile(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
	}
	return
}

func FileActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	action := ps.ByName("action")
	val, err := handleFileAction(c, r, id, action)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
	}
	return
}
