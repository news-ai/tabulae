package search

import (
	"errors"

	"gopkg.in/olivere/elastic.v3"
)

var (
	elasticClient *elastic.Client
)

func InitializeElasticSearch() error {
	err := errors.New("")
	elasticClient, err = elastic.NewClient(elastic.SetURL("https://search.newsai.org"))
	if err != nil {
		return err
	}
	return nil
}
