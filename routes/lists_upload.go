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

	"github.com/news-ai/tabulae/files"
	"github.com/news-ai/tabulae/permissions"

	"github.com/gorilla/mux"
)

// State can be some kind of random generated hash string.
// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
func randToken() string {
	b := make([]byte, 5)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Handler for when the user wants to perform an action on the lists
func MediaListActionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		w.Header().Set("Content-Type", "application/json")
		c := appengine.NewContext(r)

		user, err := GetUser(r)
		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		vars := mux.Vars(r)
		listId, ok := vars["id"]
		if !ok {
			log.Errorf(c, "%v", err)
			permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "List ID missing")
			return
		}

		userId := strconv.FormatInt(user.Id, 10)

		file, handler, err := r.FormFile("file")

		if err != nil {
			log.Errorf(c, "%v", err)
			permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		log.Debugf(c, "%v", file)
		log.Debugf(c, "%v", handler)

		noSpaceFileName := ""
		if handler.Filename != "" {
			noSpaceFileName = strings.Replace(handler.Filename, " ", "", -1)
		}
		fileName := strings.Join([]string{userId, listId, randToken(), noSpaceFileName}, "-")
		log.Debugf(c, "%v", fileName)

		val, err := files.UploadFile(r, fileName, file, userId, listId, handler.Header.Get("Content-Type"))
		if err != nil {
			log.Errorf(c, "%v", err)
			permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			log.Errorf(c, "%v", err)
			permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
		}
		return
	}
	permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "Method not implemented")
}
