package middleware

import (
	"net/http"
	"strings"

	"appengine"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

func UpdateOrCreateUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	c := appengine.NewContext(r)
	email, err := auth.GetCurrentUserEmail(r)
	if err != nil && !strings.Contains(r.URL.Path, "/api/auth/") {
		w.Header().Set("Content-Type", "application/json")
		ReturnError(w, http.StatusUnauthorized, "Not logged in", "Please login "+utils.APIURL+"/auth/google")
		return
	} else {
		if email != "" {
			userDetails, _ := auth.GetUserDetails(r)
			models.NewOrUpdateUser(c, r, email, userDetails)
		}
	}
	next(w, r)
}
