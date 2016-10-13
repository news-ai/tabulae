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
			return err
		}
	}
	return nil
}

type TwitterProfile struct {
	Type string `json:"type"`

	ID          int    `json:"id"`
	IDStr       string `json:"id_str"`
	Name        string `json:"name"`
	ScreenName  string `json:"screen_name"`
	Location    string `json:"location"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Entities    struct {
		Description struct {
			Urls []interface{} `json:"urls"`
		} `json:"description"`
	} `json:"entities"`
	Protected                      bool   `json:"protected"`
	FollowersCount                 int    `json:"followers_count"`
	FriendsCount                   int    `json:"friends_count"`
	ListedCount                    int    `json:"listed_count"`
	CreatedAt                      string `json:"created_at"`
	FavouritesCount                int    `json:"favourites_count"`
	UtcOffset                      int    `json:"utc_offset"`
	TimeZone                       string `json:"time_zone"`
	GeoEnabled                     bool   `json:"geo_enabled"`
	Verified                       bool   `json:"verified"`
	StatusesCount                  int    `json:"statuses_count"`
	Lang                           string `json:"lang"`
	ContributorsEnabled            bool   `json:"contributors_enabled"`
	IsTranslator                   bool   `json:"is_translator"`
	IsTranslationEnabled           bool   `json:"is_translation_enabled"`
	ProfileBackgroundColor         string `json:"profile_background_color"`
	ProfileBackgroundImageURL      string `json:"profile_background_image_url"`
	ProfileBackgroundImageURLHTTPS string `json:"profile_background_image_url_https"`
	ProfileBackgroundTile          bool   `json:"profile_background_tile"`
	ProfileImageURL                string `json:"profile_image_url"`
	ProfileImageURLHTTPS           string `json:"profile_image_url_https"`
	ProfileBannerURL               string `json:"profile_banner_url"`
	ProfileLinkColor               string `json:"profile_link_color"`
	ProfileSidebarBorderColor      string `json:"profile_sidebar_border_color"`
	ProfileSidebarFillColor        string `json:"profile_sidebar_fill_color"`
	ProfileTextColor               string `json:"profile_text_color"`
	ProfileUseBackgroundImage      bool   `json:"profile_use_background_image"`
	HasExtendedProfile             bool   `json:"has_extended_profile"`
	DefaultProfile                 bool   `json:"default_profile"`
	DefaultProfileImage            bool   `json:"default_profile_image"`
	Following                      bool   `json:"following"`
	FollowRequestSent              bool   `json:"follow_request_sent"`
	Notifications                  bool   `json:"notifications"`
	Username                       string `json:"Username"`
}

func (tp *TwitterProfile) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(tp, k, v)
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

func searchTwitterProfile(c context.Context, elasticQuery interface{}, username string) (TwitterProfile, error) {
	hits, err := elasticTwitterUser.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return TwitterProfile{}, err
	}

	twitterProfileHits := hits.Hits

	if len(twitterProfileHits) == 0 {
		return TwitterProfile{}, errors.New("No Twitter profile for this username")
	}

	rawTwitterProfile := twitterProfileHits[0].Source.Data
	rawMap := rawTwitterProfile.(map[string]interface{})
	twitterProfile := TwitterProfile{}
	err = twitterProfile.FillStruct(rawMap)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	if twitterProfile.Username != username {
		return TwitterProfile{}, errors.New("No Twitter profile for this username")
	}

	twitterProfile.Type = "twitterprofiles"

	return twitterProfile, nil
}

func SearchProfileByUsername(c context.Context, r *http.Request, username string) (TwitterProfile, error) {
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
