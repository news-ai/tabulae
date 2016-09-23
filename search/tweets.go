package search

import (
	"net/http"
	// "net/url"
	// "strconv"

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

func searchTweet(c context.Context, elasticQuery elastic.ElasticQuery) ([]Tweet, error) {
	hits, err := elasticTweet.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Tweet{}, err
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

		tweet.Type = "tweets"
		tweets = append(tweets, tweet)
	}

	return tweets, nil
}

func SearchTweetsByContactId(c context.Context, r *http.Request, contactId int64) ([]Tweet, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticContactIdQuery := ElasticContactIdQuery{}
	elasticContactIdQuery.Term.ContactId = contactId

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticContactIdQuery)

	return searchTweet(c, elasticQuery)
}