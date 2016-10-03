package search

import (
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
	elasticFeed *elastic.Elastic
)

type Feed struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdat"`

	Title         string `json:"title"`
	Author        string `json:"author"`
	Url           string `json:"url"`
	Summary       string `json:"summary"`
	FeedURL       string `json:"feedurl"`
	PublicationId int64  `json:"publicationid"`

	Text       string `json:"text"`
	TweetId    int64  `json:"tweetid"`
	TweetIdStr string `json:"tweetidstr"`
	Username   string `json:"username"`
}

func (f *Feed) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(f, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchFeed(c context.Context, elasticQuery interface{}, contacts []models.Contact, feedUrls []models.Feed) ([]Feed, error) {
	hits, err := elasticFeed.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Feed{}, err
	}

	feedsMap := map[string]bool{}
	for i := 0; i < len(feedUrls); i++ {
		feedsMap[strings.ToLower(feedUrls[i].FeedURL)] = true
	}

	twitterUsernamesMap := map[string]bool{}
	for i := 0; i < len(contacts); i++ {
		twitterUsernamesMap[strings.ToLower(contacts[i].Twitter)] = true
	}

	feedHits := hits.Hits
	feeds := []Feed{}
	for i := 0; i < len(feedHits); i++ {
		rawFeed := feedHits[i].Source.Data
		rawMap := rawFeed.(map[string]interface{})
		feed := Feed{}
		err := feed.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		feed.Type = strings.ToLower(feed.Type)
		feed.Type += "s"

		if feed.FeedURL != "" {
			if _, ok := feedsMap[strings.ToLower(feed.FeedURL)]; !ok {
				continue
			}
		} else {
			if _, ok := twitterUsernamesMap[strings.ToLower(feed.Username)]; !ok {
				continue
			}
		}

		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func SearchFeedForContacts(c context.Context, r *http.Request, contacts []models.Contact, feeds []models.Feed) ([]Feed, error) {
	// If twitter username or feeds are empty return right away
	if len(contacts) == 0 && len(feeds) == 0 {
		return []Feed{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticFilterWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	for i := 0; i < len(contacts); i++ {
		if contacts[i].Twitter != "" {
			elasticUsernameQuery := ElasticUsernameQuery{}
			elasticUsernameQuery.Term.Username = strings.ToLower(contacts[i].Twitter)
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticUsernameQuery)
		}
	}

	for i := 0; i < len(feeds); i++ {
		elasticFeedUrlQuery := ElasticFeedUrlQuery{}
		elasticFeedUrlQuery.Match.FeedURL = strings.ToLower(feeds[i].FeedURL)
		elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticFeedUrlQuery)
	}

	if len(elasticQuery.Query.Bool.Should) == 0 {
		return []Feed{}, nil
	}

	elasticQuery.Query.Bool.MinimumShouldMatch = "50%"
	elasticQuery.MinScore = 0.2

	elasticCreatedAtQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtQuery)

	return searchFeed(c, elasticQuery, contacts, feeds)
}
