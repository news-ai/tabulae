package search

import (
	"gopkg.in/olivere/elastic.v3"
)

type AgencySearch struct {
	Id   int64
	Name string
}

func SearchAgency(search string) (interface{}, error) {
	termQuery := elastic.NewTermQuery("data.Name", search)
	searchResult, err := elasticClient.Search().Index("Agency").Query(termQuery).Do()

	if err != nil {
		return nil, err
	}
}
