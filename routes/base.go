package routes

import (
	"net/http"

	"google.golang.org/appengine"

	"golang.org/x/net/context"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/permissions"
)

var resourcesHandlers map[string](func(context.Context, http.ResponseWriter, *http.Request) (interface{}, error))

// Handler for when there is a key present after /users/<id> route.
func NotFoundHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	permissions.ReturnError(w, http.StatusNotFound, "An unknown error occurred while trying to process this request.", "Not Found")
	return
}

func InitializeResourceHandlers() {
	resourcesHandlers = make(map[string](func(context.Context, http.ResponseWriter, *http.Request) (interface{}, error)))
	// resourcesHandlers["users"] = handleUsers
	// resourcesHandlers["agencies"] = handleAgencies
	// resourcesHandlers["publications"] = handlePublications
	// resourcesHandlers["contacts"] = handleContacts
	// resourcesHandlers["files"] = handleFiles
	// resourcesHandlers["lists"] = handleMediaLists
	// resourcesHandlers["emails"] = handleEmails
}

func ResourcesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	c := appengine.NewContext(r)

	w.Header().Set("Content-Type", "application/json")
	resource := ps.ByName("resource")
	val, err := resourcesHandlers[resource](c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Agency handling error", err.Error())
	}
	return
}
