package middleware

import (
	"net/http"

	"appengine"

	"github.com/news-ai/tabulae/models"
)

func UpdateOrCreateUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	c := appengine.NewContext(r)
	models.NewOrUpdateUser(c)
	next(w, r)
}
