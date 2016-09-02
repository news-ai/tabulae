package search

import (
	"fmt"
	"os"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	baseURL = "https://%v:%v@search.newsai.org"
)

func InitializeElasticSearch() {
	baseURL = fmt.Sprintf(baseURL, os.Getenv("ELASTIC_USER"), os.Getenv("ELASTIC_PASS"))

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
}
