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
	elasticTweet *elastic.Elastic
)

type Tweet struct {
	Type string `json:"type"`

	Text      string    `json:"text"`
	TweetId   int64     `json:"tweetid"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdat"`
}

func (t *Tweet) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(t, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchTweet(c context.Context, elasticQuery interface{}, usernames []string) ([]Tweet, error) {
	hits, err := elasticTweet.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Tweet{}, err
	}

	usernamesMap := map[string]bool{}
	for i := 0; i < len(usernames); i++ {
		usernamesMap[strings.ToLower(usernames[i])] = true
	}

	tweetHits := hits.Hits
	tweets := []Tweet{}
	for i := 0; i < len(tweetHits); i++ {
		rawTweet := tweetHits[i].Source.Data
		rawMap := rawTweet.(map[string]interface{})
		tweet := Tweet{}
		err := tweet.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		if _, ok := usernamesMap[strings.ToLower(tweet.Username)]; !ok {
			continue
		}

		tweet.Type = "tweets"
		tweets = append(tweets, tweet)
	}

	return tweets, nil
}

func SearchTweetsByUsername(c context.Context, r *http.Request, username string) ([]Tweet, error) {
	if username == "" {
		return []Tweet{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticFilterWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUsernameQuery := ElasticUsernameQuery{}
	elasticUsernameQuery.Term.Username = strings.ToLower(username)

	elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticUsernameQuery)

	elasticQuery.Query.Bool.MinimumShouldMatch = "100%"

	elasticCreatedAtQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtQuery)

	return searchTweet(c, elasticQuery, []string{username})
}

func SearchTweetsByUsernames(c context.Context, r *http.Request, usernames []string) ([]Tweet, error) {
	if len(usernames) == 0 {
		return []Tweet{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticFilterWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	for i := 0; i < len(usernames); i++ {
		elasticUsernameQuery := ElasticUsernameMatchQuery{}
		elasticUsernameQuery.Match.Username = strings.ToLower(usernames[i])
		elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticUsernameQuery)
	}

	elasticQuery.Query.Bool.MinimumShouldMatch = "0"
	elasticQuery.MinScore = 0

	elasticCreatedAtQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtQuery)

	return searchTweet(c, elasticQuery, usernames)
}
