package controllers

import (
	"net/http"

	"google.golang.org/appengine/datastore"

	gcontext "github.com/gorilla/context"
)

func constructQuery(query *datastore.Query, r *http.Request) *datastore.Query {
	order := gcontext.Get(r, "order").(string)
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	if order != "" {
		query = query.Order(order)
	}

	return query.Limit(limit).Offset(offset)
}
