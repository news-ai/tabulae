package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/news-ai/tabulae/permissions"
)

// Handler for when there is a key present after /users/<id> route.
func NotFoundHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	permissions.ReturnError(w, http.StatusNotFound, "An unknown error occurred while trying to process this request.", "Not Found")
	return
}
