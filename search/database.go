package search

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticContactDatabase *elastic.Elastic
)

type EnhanceResponse struct {
	Data interface{} `json:"data"`
}

type EnhanceFullContactProfileResponse struct {
	Data interface{} `json:"data"`
}

type DatabaseResponse struct {
	Email string      `json:"email"`
	Data  interface{} `json:"data"`
}

func searchESContactsDatabase(c context.Context, elasticQuery elastic.ElasticQuery) (interface{}, int, error) {
	hits, err := elasticContactDatabase.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	contactHits := hits.Hits
	var contacts []interface{}
	for i := 0; i < len(contactHits); i++ {
		rawContact := contactHits[i].Source.Data
		contactData := DatabaseResponse{
			Email: contactHits[i].ID,
			Data:  rawContact,
		}
		contacts = append(contacts, contactData)
	}

	return contacts, len(contactHits), nil
}

func SearchCompanyDatabase(c context.Context, r *http.Request, url string) (interface{}, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/company/" + url

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	var enhanceResponse EnhanceResponse
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	return enhanceResponse.Data, nil
}

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

	var enhanceResponse EnhanceFullContactProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	return enhanceResponse.Data, nil
}

func SearchESContactsDatabase(c context.Context, r *http.Request) (interface{}, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	return searchESContactsDatabase(c, elasticQuery)
}
