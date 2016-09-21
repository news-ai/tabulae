package search

import (
	elastic "github.com/news-ai/elastic-appengine"
)

var (
	baseURL = "https://search.newsai.org"
)

type ElasticCreatedByQuery struct {
	Term struct {
		CreatedBy int64 `json:"data.CreatedBy"`
	} `json:"term"`
}

type ElasticListIdQuery struct {
	Term struct {
		ListId int64 `json:"data.ListId"`
	} `json:"term"`
}

type ElasticContactIdQuery struct {
	Term struct {
		ContactId int64 `json:"data.ContactId"`
	} `json:"term"`
}

type ElasticPublicationIdQuery struct {
	Term struct {
		PublicationId int64 `json:"data.PublicationId"`
	} `json:"term"`
}

type ElasticFeedUrlQuery struct {
	Match struct {
		FeedURL string `json:"data.FeedURL"`
	} `json:"match"`
}

type ElasticSortDataPublishDateQuery struct {
	DataPublishDate struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.PublishDate"`
}

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

	headlineElastic := elastic.Elastic{}
	headlineElastic.BaseURL = baseURL
	headlineElastic.Index = "headlines"
	headlineElastic.Type = "headline"
	elasticHeadline = &headlineElastic
}
