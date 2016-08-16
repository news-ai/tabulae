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

	"github.com/news-ai/tabulae/controllers"
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
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	user, err := GetUser(r)
	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
		return
	}

	userId := strconv.FormatInt(user.Id, 10)

	// Get Id and Action
	vars := mux.Vars(r)
	listId, idOk := vars["id"]
	action, actionOk := vars["action"]

	if idOk && actionOk {
		switch r.Method {
		case "GET":
			if action == "contacts" {
				limit, offset, err := GetPagination(r)

				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "Media List handling error", err.Error())
					return
				}

				val, err := controllers.GetContactsForList(c, r, listId, limit, offset)

				if err == nil {
					err = json.NewEncoder(w).Encode(val)
				}

				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "Media List handling error", err.Error())
					return
				}
				return
			}
		case "POST":
			if action == "upload" {
				file, handler, err := r.FormFile("file")

				if err != nil {
					log.Errorf(c, "%v", err)
					permissions.ReturnError(w, http.StatusInternalServerError, "List handling error", err.Error())
					return
				}

				noSpaceFileName := ""
				if handler.Filename != "" {
					noSpaceFileName = strings.Replace(handler.Filename, " ", "", -1)
				}
				fileName := strings.Join([]string{userId, listId, randToken(), noSpaceFileName}, "-")

				val, err := files.UploadFile(r, fileName, file, userId, listId, handler.Header.Get("Content-Type"))
				if err != nil {
					log.Errorf(c, "%v", err)
					permissions.ReturnError(w, http.StatusInternalServerError, "List handling error", err.Error())
					return
				}

				if err == nil {
					err = json.NewEncoder(w).Encode(val)
				}

				if err != nil {
					log.Errorf(c, "%v", err)
					permissions.ReturnError(w, http.StatusInternalServerError, "List handling error", err.Error())
					return
				}
				return
			}
		default:
			permissions.ReturnError(w, http.StatusInternalServerError, "List handling error", "Method not implemented")
			return
		}
		permissions.ReturnError(w, http.StatusInternalServerError, "List handling error", "Action not valid")
		return
	}
	permissions.ReturnError(w, http.StatusInternalServerError, "List handling error", "Id or action not valid")
	return
}
