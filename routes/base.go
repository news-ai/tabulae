package routes

import (
	"net/http"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/permissions"
)

var resourcesHandlers map[string](func(context.Context, http.ResponseWriter, *http.Request) (interface{}, error))

func baseResponseHandler(val interface{}, included interface{}, count int, err error, r *http.Request) (models.BaseResponse, error) {
	response := models.BaseResponse{}
	response.Data = val
	response.Included = included
	response.Count = count
	response.Next = gcontext.Get(r, "next").(string)
	return response, err
}

func baseSingleResponseHandler(val interface{}, included interface{}, err error) (models.BaseSingleResponse, error) {
	response := models.BaseSingleResponse{}
	response.Data = val
	response.Included = included
	return response, err
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	permissions.ReturnError(w, http.StatusNotFound, "An unknown error occurred while trying to process this request.", "Not Found")
	return
}

// Handler for when there is a key present after /users/<id> route.
func NotFoundHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	permissions.ReturnError(w, http.StatusNotFound, "An unknown error occurred while trying to process this request.", "Not Found")
	return
}
