package search

import (
	elastic "github.com/news-ai/elastic-appengine"

	"github.com/news-ai/api/search"
)

func InitializeElasticSearch() {
	agencyElastic := elastic.Elastic{}
	agencyElastic.BaseURL = search.NewBaseURL
	agencyElastic.Index = "agencies"
	agencyElastic.Type = "agency"
	elasticAgency = &agencyElastic

	publicationElastic := elastic.Elastic{}
	publicationElastic.BaseURL = search.NewBaseURL
	publicationElastic.Index = "publications"
	publicationElastic.Type = "publication"
	elasticPublication = &publicationElastic

	contactElastic := elastic.Elastic{}
	contactElastic.BaseURL = search.NewBaseURL
	contactElastic.Index = "contacts"
	contactElastic.Type = "contact"
	elasticContact = &contactElastic

	listTimeseriesElastic := elastic.Elastic{}
	listTimeseriesElastic.BaseURL = search.NewBaseURL
	listTimeseriesElastic.Index = "lists"
	listTimeseriesElastic.Type = "list"
	elasticList = &listTimeseriesElastic

	emailLogElastic := elastic.Elastic{}
	emailLogElastic.BaseURL = search.NewBaseURL
	emailLogElastic.Index = "emails"
	emailLogElastic.Type = "log"
	elasticEmailLog = &emailLogElastic

	emailTimeseriesElastic := elastic.Elastic{}
	emailTimeseriesElastic.BaseURL = search.NewBaseURL
	emailTimeseriesElastic.Index = "timeseries"
	emailTimeseriesElastic.Type = "useremail2"
	elasticEmailTimeseries = &emailTimeseriesElastic

	emailsElastic := elastic.Elastic{}
	emailsElastic.BaseURL = search.NewBaseURL
	emailsElastic.Index = "emails1"
	emailsElastic.Type = "email"
	elasticEmails = &emailsElastic

	emailsElasticCampaign := elastic.Elastic{}
	emailsElasticCampaign.BaseURL = search.NewBaseURL
	emailsElasticCampaign.Index = "emails"
	emailsElasticCampaign.Type = "campaign1"
	elasticEmailCampaign = &emailsElasticCampaign
}
