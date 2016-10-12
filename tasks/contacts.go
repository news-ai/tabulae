package tasks

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"

	"github.com/news-ai/web/errors"
)

func SocialUsernameInvalid(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// If successful
	w.WriteHeader(200)
	return
}
