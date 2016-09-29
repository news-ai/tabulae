package search

import (
	"net/http"
	// "net/url"
	// "strconv"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/models"
)

type Feed struct {
}

func SearchFeedForContact(c context.Context, r *http.Request, contact models.Contact, feeds []models.Feed) ([]Feed, error) {
	feed := []Feed{}

	return feed, nil
}
