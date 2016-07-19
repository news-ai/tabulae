package routes

import (
	"net/http"

	"github.com/news-ai/tabulae/middleware"
)

// Handler for when the user wants all the users.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	middleware.ReturnError(w, http.StatusInternalServerError, "Publication handling error", "")
}
