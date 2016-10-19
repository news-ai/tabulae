package search

import (
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticInstagramTimeseries *elastic.Elastic
	elasticTwitterTimeseries   *elastic.Elastic
)

func searchTwitterTimeseries(c context.Context, elasticQuery interface{}) (interface{}, error) {
	hits, err := elasticTwitterTimeseries.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	profileHits := hits.Hits

	if len(profileHits) == 0 {
		log.Infof(c, "%v", profileHits)
		return nil, errors.New("No Twitter profile for this username")
	}

	var interfaceSlice = make([]interface{}, len(profileHits))

	for i := 0; i < len(profileHits); i++ {
		interfaceSlice[i] = profileHits[i].Source.Data
	}

	return interfaceSlice, nil
}

func searchInstagramTimeseries(c context.Context, elasticQuery interface{}) (interface{}, error) {
	hits, err := elasticInstagramTimeseries.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	profileHits := hits.Hits

	if len(profileHits) == 0 {
		log.Infof(c, "%v", profileHits)
		return nil, errors.New("No Instagram profile for this username")
	}

	var interfaceSlice = make([]interface{}, len(profileHits))

	for i := 0; i < len(profileHits); i++ {
		interfaceSlice[i] = profileHits[i].Source.Data
	}

	return interfaceSlice, nil
}

func SearchInstagramTimeseriesByUsername(c context.Context, r *http.Request, username string) (interface{}, error) {
	if username == "" {
		return nil, errors.New("Contact does not have a instagram username")
	}

	offset := 0
	limit := 31

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUsernameQuery := ElasticUsernameQuery{}
	elasticUsernameQuery.Term.Username = strings.ToLower(username)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUsernameQuery)

	elasticCreatedAtDateQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtDateQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtDateQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtDateQuery)

	return searchInstagramTimeseries(c, elasticQuery)
}

func SearchTwitterTimeseriesByUsername(c context.Context, r *http.Request, username string) (interface{}, error) {
	if username == "" {
		return nil, errors.New("Contact does not have a twitter username")
	}

	offset := 0
	limit := 31

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUsernameQuery := ElasticUsernameQuery{}
	elasticUsernameQuery.Term.Username = strings.ToLower(username)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUsernameQuery)

	elasticCreatedAtDateQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtDateQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtDateQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtDateQuery)

	return searchTwitterTimeseries(c, elasticQuery)
}
