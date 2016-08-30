package search

import (
	elastic "github.com/news-ai/elastic-appengine"
)

var (
	baseURL = "https://search.newsai.org"
)

func InitializeElasticSearch() {
	agencyElastic := elastic.Elastic{}
	agencyElastic.BaseURL = baseURL
	agencyElastic.ResourceType = "agencies"
	elasticAgency = &agencyElastic

	publicationElastic := elastic.Elastic{}
	publicationElastic.BaseURL = baseURL
	publicationElastic.ResourceType = "publications"
	elasticPublication = &publicationElastic
}
