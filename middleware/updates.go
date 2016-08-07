package middleware

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
	"github.com/news-ai/tabulae/utils"
)

func UpdateOrCreateUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	c := appengine.NewContext(r)
	email, err := auth.GetCurrentUserEmail(r)
	if err != nil && !strings.Contains(r.URL.Path, "/api/auth") && !strings.Contains(r.URL.Path, "/static") {
		w.Header().Set("Content-Type", "application/json")
		permissions.ReturnError(w, http.StatusUnauthorized, "Authentication Required", "Please login "+utils.APIURL+"/auth/google")
		return
	} else {
		if email != "" {
			userDetails, _ := auth.GetUserDetails(r)
			controllers.NewOrUpdateUser(c, r, email, userDetails)
		}
	}
	next(w, r)
}
