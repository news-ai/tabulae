package search

import (
	elastic "github.com/news-ai/elastic-appengine"
)

var (
	baseURL    = "https://search.newsai.org"
	newBaseURL = "https://search1.newsai.org"
)

type ElasticMGetQuery struct {
	Ids []string `json:"ids"`
}

type ElasticCreatedByQuery struct {
	Term struct {
		CreatedBy int64 `json:"data.CreatedBy"`
	} `json:"term"`
}

type ElasticClientQuery struct {
	Term struct {
		Client string `json:"data.Client"`
	} `json:"match"`
}

type ElasticTagQuery struct {
	Term struct {
		Tag string `json:"data.Tags"`
	} `json:"match"`
}

type ElasticAllQuery struct {
	Term struct {
		All string `json:"_all"`
	} `json:"match"`
}

type ElasticArchivedQuery struct {
	Term struct {
		Archived bool `json:"data.Archived"`
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

type ElasticUsernameQuery struct {
	Term struct {
		Username string `json:"data.Username"`
	} `json:"term"`
}

type ElasticInstagramUsernameQuery struct {
	Term struct {
		InstagramUsername string `json:"data.InstagramUsername"`
	} `json:"term"`
}

type ElasticFeedUsernameQuery struct {
	Term struct {
		Type     string `json:"data.Type"`
		Username string `json:"data.Username"`
	} `json:"term"`
}

type ElasticUsernameMatchQuery struct {
	Match struct {
		Username string `json:"data.Username"`
	} `json:"match"`
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

type ElasticSortDataCreatedAtQuery struct {
	DataCreatedAt struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.CreatedAt"`
}

func InitializeElasticSearch() {
	agencyElastic := elastic.Elastic{}
	agencyElastic.BaseURL = newBaseURL
	agencyElastic.Index = "agencies"
	agencyElastic.Type = "agency"
	elasticAgency = &agencyElastic

	publicationElastic := elastic.Elastic{}
	publicationElastic.BaseURL = newBaseURL
	publicationElastic.Index = "publications"
	publicationElastic.Type = "publication"
	elasticPublication = &publicationElastic

	contactElastic := elastic.Elastic{}
	contactElastic.BaseURL = newBaseURL
	contactElastic.Index = "contacts"
	contactElastic.Type = "contact"
	elasticContact = &contactElastic

	headlineElastic := elastic.Elastic{}
	headlineElastic.BaseURL = newBaseURL
	headlineElastic.Index = "headlines"
	headlineElastic.Type = "headline"
	elasticHeadline = &headlineElastic

	tweetElastic := elastic.Elastic{}
	tweetElastic.BaseURL = newBaseURL
	tweetElastic.Index = "tweets"
	tweetElastic.Type = "tweet"
	elasticTweet = &tweetElastic

	twitterUserElastic := elastic.Elastic{}
	twitterUserElastic.BaseURL = newBaseURL
	twitterUserElastic.Index = "tweets"
	twitterUserElastic.Type = "user"
	elasticTwitterUser = &twitterUserElastic

	feedElastic := elastic.Elastic{}
	feedElastic.BaseURL = newBaseURL
	feedElastic.Index = "feeds"
	feedElastic.Type = "feed"
	elasticFeed = &feedElastic

	instagramElastic := elastic.Elastic{}
	instagramElastic.BaseURL = newBaseURL
	instagramElastic.Index = "instagrams"
	instagramElastic.Type = "instagram"
	elasticInstagram = &instagramElastic

	instagramUserElastic := elastic.Elastic{}
	instagramUserElastic.BaseURL = newBaseURL
	instagramUserElastic.Index = "instagrams"
	instagramUserElastic.Type = "user"
	elasticInstagramUser = &instagramUserElastic

	instagramTimeseriesElastic := elastic.Elastic{}
	instagramTimeseriesElastic.BaseURL = newBaseURL
	instagramTimeseriesElastic.Index = "timeseries"
	instagramTimeseriesElastic.Type = "instagram"
	elasticInstagramTimeseries = &instagramTimeseriesElastic

	twitterTimeseriesElastic := elastic.Elastic{}
	twitterTimeseriesElastic.BaseURL = newBaseURL
	twitterTimeseriesElastic.Index = "timeseries"
	twitterTimeseriesElastic.Type = "twitter"
	elasticTwitterTimeseries = &twitterTimeseriesElastic

	listTimeseriesElastic := elastic.Elastic{}
	listTimeseriesElastic.BaseURL = newBaseURL
	listTimeseriesElastic.Index = "lists"
	listTimeseriesElastic.Type = "list"
	elasticList = &listTimeseriesElastic
}
