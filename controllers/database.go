package controllers

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/database"
)

func GetDatabases(c context.Context, r *http.Request) (interface{}, interface{}, int, error) {
	contacts, included, count, err := database.GetAllContacts(c, r)
	if err != nil {
		return nil, nil, 0, err
	}
	return contacts, included, count, nil
}
