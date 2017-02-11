package tasks

import (
	"net/http"
	// "google.golang.org/appengine"
)

func RemoveUnsentEmailsHandler(w http.ResponseWriter, r *http.Request) {
	// c := appengine.NewContext(r)

	w.WriteHeader(200)
	return
}
