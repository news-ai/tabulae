package search

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	"gopkg.in/olivere/elastic.v3"
)

type AgencySearch struct {
	Id   int64
	Name string
}

func SearchAgency(c context.Context, search string) (interface{}, error) {
	termQuery := elastic.NewTermQuery("data.Name", search)
	searchResult, err := elasticClient.Search().Index("Agency").Query(termQuery).Do()

	log.Infof(c, "%v", searchResult)

	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	log.Infof(c, "%v", searchResult)

	return nil, nil
}
