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
	// Basic authentication
	apiKey, _, _ := r.BasicAuth()
	apiKeyValid := false
	if apiKey != "" {
		apiKeyValid = auth.BasicAuthLogin(w, r, apiKey)
	}

	c := appengine.NewContext(r)
	email, err := auth.GetCurrentUserEmail(r)
	if err != nil && !strings.Contains(r.URL.Path, "/api/auth") && !strings.Contains(r.URL.Path, "/static") && !apiKeyValid {
		w.Header().Set("Content-Type", "application/json")
		permissions.ReturnError(w, http.StatusUnauthorized, "Authentication Required", "Please login "+utils.APIURL+"/auth/google")
		return
	} else {
		if email != "" {
			controllers.AddUserToContext(c, r, email)
		}
	}

	next(w, r)
}
