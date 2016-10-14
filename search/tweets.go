package search

import (
	"errors"
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
	elasticTweet       *elastic.Elastic
	elasticTwitterUser *elastic.Elastic
)

type Tweet struct {
	Type string `json:"type"`

	Text       string    `json:"text"`
	TweetId    int64     `json:"tweetid"`
	TweetIdStr string    `json:"tweetidstr"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"createdat"`
}

func (t *Tweet) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(t, k, v)
		if err != nil {
			// return err
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

func searchTwitterProfile(c context.Context, elasticQuery interface{}, username string) (interface{}, error) {
	hits, err := elasticTwitterUser.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return TwitterProfile{}, err
	}

	twitterProfileHits := hits.Hits

	if len(twitterProfileHits) == 0 {
		log.Infof(c, "%v", twitterProfileHits)
		return TwitterProfile{}, errors.New("No Twitter profile for this username")
	}

	return twitterProfileHits[0].Source.Data, nil
}

func SearchProfileByUsername(c context.Context, r *http.Request, username string) (interface{}, error) {
	if username == "" {
		return TwitterProfile{}, nil
	}

	offset := 0
	limit := 1

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUsernameQuery := ElasticUsernameQuery{}
	elasticUsernameQuery.Term.Username = strings.ToLower(username)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUsernameQuery)

	return searchTwitterProfile(c, elasticQuery, username)
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
		if usernames[i] != "" {
			elasticUsernameQuery := ElasticUsernameMatchQuery{}
			elasticUsernameQuery.Match.Username = strings.ToLower(usernames[i])
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticUsernameQuery)
		}
	}

	if len(elasticQuery.Query.Bool.Should) == 0 {
		return []Tweet{}, nil
	}

	elasticQuery.Query.Bool.MinimumShouldMatch = "0"
	elasticQuery.MinScore = 0

	elasticCreatedAtQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtQuery)

	return searchTweet(c, elasticQuery, usernames)
}
