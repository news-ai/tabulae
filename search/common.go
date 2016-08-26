package search

import (
	"gopkg.in/olivere/elastic.v3"
)

var (
	ElasticClient *elastic.Client
)

func InitializeElasticSearch() error {
	ElasticClient, err := elastic.NewClient(elastic.SetURL("https://search.newsai.org"))
	if err != nil {
		return err
	}
}
