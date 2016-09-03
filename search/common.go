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
	agencyElastic.Index = "agencies"
	agencyElastic.Type = "agency"
	elasticAgency = &agencyElastic

	publicationElastic := elastic.Elastic{}
	publicationElastic.BaseURL = baseURL
	publicationElastic.Index = "publications"
	publicationElastic.Type = "publication"
	elasticPublication = &publicationElastic

	contactElastic := elastic.Elastic{}
	contactElastic.BaseURL = baseURL
	contactElastic.Index = "contacts"
	contactElastic.Type = "contact"
	elasticContact = &contactElastic
}
