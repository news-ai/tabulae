package search

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticHeadline *elastic.Elastic
)

type Headline struct {
	Type string `json:"type"`

	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Url         string    `json:"url"`
	Categories  []string  `json:"categories"`
	PublishDate time.Time `json:"createdat"`
	Summary     string    `json:"summary"`
	FeedURL     string    `json:"feedurl"`

	PublicationId int64 `json:"publicationid"`
}

func (h *Headline) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(h, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchHeadline(c context.Context, elasticQuery interface{}, feedUrls []models.Feed) ([]Headline, error) {
	hits, err := elasticHeadline.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Headline{}, err
	}

	feedsMap := map[string]bool{}
	for i := 0; i < len(feedUrls); i++ {
		feedsMap[strings.ToLower(feedUrls[i].FeedURL)] = true
	}

	headlineHits := hits.Hits
	headlines := []Headline{}
	for i := 0; i < len(headlineHits); i++ {
		rawHeadline := headlineHits[i].Source.Data
		rawMap := rawHeadline.(map[string]interface{})
		headline := Headline{}
		err := headline.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		headline.Type = "headlines"

		if headline.FeedURL != "" {
			if _, ok := feedsMap[strings.ToLower(headline.FeedURL)]; !ok {
				continue
			}
		}

		headlines = append(headlines, headline)
	}

	return headlines, nil
}

func SearchHeadlinesByResourceId(c context.Context, r *http.Request, feeds []models.Feed) ([]Headline, error) {
	if len(feeds) == 0 {
		return []Headline{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticFilterWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	for i := 0; i < len(feeds); i++ {
		elasticFeedUrlQuery := ElasticFeedUrlQuery{}
		elasticFeedUrlQuery.Match.FeedURL = strings.ToLower(feeds[i].FeedURL)
		elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticFeedUrlQuery)
	}

	if len(elasticQuery.Query.Bool.Should) == 0 {
		return []Headline{}, nil
	}

	minMatch := "100%"
	if len(elasticQuery.Query.Bool.Should) > 1 {
		approxMatch := float64(100 / len(elasticQuery.Query.Bool.Should))
		minMatch = fmt.Sprint(approxMatch) + "%"
	}

	elasticQuery.Query.Bool.MinimumShouldMatch = minMatch
	elasticQuery.MinScore = 0.6

	elasticPublishDateQuery := ElasticSortDataPublishDateQuery{}
	elasticPublishDateQuery.DataPublishDate.Order = "desc"
	elasticPublishDateQuery.DataPublishDate.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticPublishDateQuery)

	return searchHeadline(c, elasticQuery, feeds)
}

func SearchHeadlinesByPublicationId(c context.Context, r *http.Request, publicationId int64) ([]Headline, error) {
	if publicationId == 0 {
		return []Headline{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticPublicationIdQuery := ElasticPublicationIdQuery{}
	elasticPublicationIdQuery.Term.PublicationId = publicationId

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticPublicationIdQuery)
	elasticPublishDateQuery := ElasticSortDataPublishDateQuery{}
	elasticPublishDateQuery.DataPublishDate.Order = "desc"
	elasticPublishDateQuery.DataPublishDate.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticPublishDateQuery)

	return searchHeadline(c, elasticQuery, []models.Feed{})
}
