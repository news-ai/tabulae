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
	elasticEmailLog *elastic.Elastic
	elasticEmails   *elastic.Elastic
)

func searchEmail(c context.Context, elasticQuery interface{}) (interface{}, int, error) {
	hits, err := elasticEmails.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	log.Infof(c, "%v", hits)

	emailLogHits := []interface{}{}
	for i := 0; i < len(hits.Hits); i++ {
		emailLogHits = append(emailLogHits, hits.Hits[i].Source.Data)
	}

	return emailLogHits, len(emailLogHits), nil
}

func SearchEmailLogByEmailId(c context.Context, r *http.Request, user models.User, emailId int64) (interface{}, int, error) {
	if emailId == 0 {
		return nil, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticEmailIdQuery := ElasticEmailIdQuery{}
	elasticEmailIdQuery.Term.EmailId = emailId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmailIdQuery)

	return searchEmail(c, elasticQuery)
}

func SearchEmailLogByQuery(c context.Context, r *http.Request, user models.User, searchQuery string) (interface{}, int, error) {
	if searchQuery == "" {
		return nil, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = searchQuery

	// elasticEmailToQuery := ElasticEmailToQuery{}
	// elasticEmailToQuery.Term.To = searchQuery

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	elasticCreatedQuery := ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmail(c, elasticQuery)
}
