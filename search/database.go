package search

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticContactDatabase *elastic.Elastic
)

func SearchContactDatabase(c context.Context, r *http.Request, email string) (interface{}, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/fullcontact/" + email

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	return resp.Body, nil
}
