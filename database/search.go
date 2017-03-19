package database

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/search"
)

func GetAllContacts(c context.Context, r *http.Request) (interface{}, interface{}, int, error) {
	contacts, count, err := search.SearchESContactsDatabase(c, r)
	if err != nil {
		return nil, nil, 0, err
	}
	return contacts, nil, count, nil
}
