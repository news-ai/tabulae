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
	elasticInstagram *elastic.Elastic
)

type InstagramPost struct {
	Type string `json:"type"`

	Video       string   `json:"video"`
	Tags        []string `json:"tags"`
	Location    string   `json:"location"`
	Coordinates string   `json:"location"`
	Comments    int      `json:"comments"`
	Likes       int      `json:"likes"`

	Link  string `json:"link"`
	Image string `json:"image"`

	Caption     string `json:"caption"`
	InstagramId string `json:"instagramid"`
	Username    string `json:"Username"`

	IsDeleted bool

	CreatedAt time.Time `json:"createdat"`
}

func (ip *InstagramPost) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(ip, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchInstagramPost(c context.Context, elasticQuery interface{}, usernames []string) ([]InstagramPost, error) {
	hits, err := elasticInstagram.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []InstagramPost{}, err
	}

	usernamesMap := map[string]bool{}
	for i := 0; i < len(usernames); i++ {
		usernamesMap[strings.ToLower(usernames[i])] = true
	}

	instagramHits := hits.Hits
	instagramPosts := []InstagramPost{}
	for i := 0; i < len(instagramHits); i++ {
		rawInstagramPost := instagramHits[i].Source.Data
		rawMap := rawInstagramPost.(map[string]interface{})
		instagramPost := InstagramPost{}
		err := instagramPost.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		if _, ok := usernamesMap[strings.ToLower(instagramPost.Username)]; !ok {
			continue
		}

		instagramPost.Type = "instagrams"
		instagramPosts = append(instagramPosts, instagramPost)
	}

	return instagramPosts, nil
}

func SearchInstagramPostsByUsername(c context.Context, r *http.Request, username string) ([]InstagramPost, error) {
	if username == "" {
		return []InstagramPost{}, nil
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

	return searchInstagramPost(c, elasticQuery, []string{username})
}
