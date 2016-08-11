package middleware

import (
	"net/http"

	"github.com/news-ai/tabulae/auth"
)

func BasicAuthMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	apiKey, _, _ := r.BasicAuth()

	if apiKey != "" {
		auth.BasicAuthLogin(w, r, apiKey)
	}

	// Call the next middleware handler
	next(w, r)

	if apiKey != "" {
		auth.BasicAuthLogout(w, r)
	}
}
