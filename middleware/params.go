package middleware

import (
	"net/http"
	"strconv"

	gcontext "github.com/gorilla/context"

	"github.com/news-ai/web/utilities"
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

func GetParams(r *http.Request) (string, string, string) {
	url := utilities.StripQueryString(r.URL.String())
	searchQuery := r.URL.Query().Get("q")
	order := r.URL.Query().Get("order")
	return url, searchQuery, order
}

func ConstructNext(r *http.Request, limit int, offset int, query string, order string) string {
	url := r.URL
	q := r.URL.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset+limit))

	if query != "" {
		q.Set("query", query)
	}

	if order != "" {
		q.Set("order", order)
	}

	url.RawQuery = q.Encode()
	return url.String()
}

func AttachParameters(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	limit, offset := GetPagination(r)
	url, query, order := GetParams(r)
	nextUrl := ConstructNext(r, limit, offset, query, order)
	gcontext.Set(r, "query", query)
	gcontext.Set(r, "url", url)
	gcontext.Set(r, "order", order)
	gcontext.Set(r, "limit", limit)
	gcontext.Set(r, "offset", offset)
	gcontext.Set(r, "next", nextUrl)
	next(w, r)
}
