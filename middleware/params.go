package middleware

import (
	"net/http"
	"strconv"

	gcontext "github.com/gorilla/context"

	"github.com/news-ai/tabulae/utils"
)

func GetPagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	queryLimit := r.URL.Query().Get("limit")
	queryOffset := r.URL.Query().Get("offset")

	// check if query exists
	if len(queryLimit) != 0 {
		limit, _ = strconv.Atoi(queryLimit)
	}

	// check if offset exists
	if len(queryOffset) != 0 {
		offset, _ = strconv.Atoi(queryOffset)
	}

	// Boundary checks
	max_limit := 50
	if limit > max_limit {
		limit = max_limit
	}

	return limit, offset
}

func GetURL(r *http.Request) string {
	return utils.StripQueryString(r.URL.String())
}

func AttachParameters(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	limit, offset := GetPagination(r)
	url := GetURL(r)
	gcontext.Set(r, "url", url)
	gcontext.Set(r, "limit", limit)
	gcontext.Set(r, "offset", offset)
	next(w, r)
}
