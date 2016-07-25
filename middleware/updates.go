package middleware

import (
	"net/http"

	"appengine"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/models"
)

func UpdateOrCreateUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	c := appengine.NewContext(r)
	email, err := auth.GetCurrentUserEmail(r)
	c.Infof("%v", r.URL.Path)
	if err != nil && (r.URL.Path != "/api/auth/google" && r.URL.Path != "/api/auth/callback") {
		c.Infof("%v", r)
		w.Header().Set("Content-Type", "application/json")
		ReturnError(w, http.StatusUnauthorized, "Not logged in", "Please login http://tabulae.newsai.org/api/auth/google")
		return
	} else {
		if email != "" {
			userDetails, _ := auth.GetUserDetails(r)
			models.NewOrUpdateUser(c, r, email, userDetails)
		}
	}
	next(w, r)
}
