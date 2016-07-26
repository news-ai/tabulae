package routes

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/upload"

	"github.com/gorilla/mux"
)

// State can be some kind of random generated hash string.
// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
func randToken() string {
	b := make([]byte, 5)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Handler for when the user wants all the users.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		w.Header().Set("Content-Type", "application/json")
		c := appengine.NewContext(r)

		user, err := GetUser(r)
		if err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		vars := mux.Vars(r)
		listId, ok := vars["id"]
		if !ok {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "List ID missing")
			return
		}

		log.Debugf(c, "%v", user.Id)

		userId := strconv.FormatInt(user.Id, 10)

		file, handler, err := r.FormFile("file")
		noSpaceFileName := strings.Replace(handler.Filename, " ", "", -1)
		fileName := strings.Join([]string{userId, listId, randToken(), noSpaceFileName}, "-")

		val, err := upload.UploadFile(r, fileName, file, userId, listId, handler.Header.Get("Content-Type"))
		if err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
		}
		return
	}
	middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "Method not implemented")
}
