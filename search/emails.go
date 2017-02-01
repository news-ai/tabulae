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
)

func searchEmailLog(c context.Context, elasticQuery elastic.ElasticQuery) (interface{}, error) {
	hits, err := elasticEmailLog.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	emailLogHits := []interface{}{}
	for i := 0; i < len(hits.Hits); i++ {
		emailLogHits = append(emailLogHits, hits.Hits[i].Source.Data)
	}

	return emailLogHits, nil
}

func SearchEmailLogByEmailId(c context.Context, r *http.Request, user models.User, emailId int64) (interface{}, error) {
	if emailId == 0 {
		return nil, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticEmailIdQuery := ElasticEmailIdQuery{}
	elasticEmailIdQuery.Term.EmailId = emailId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmailIdQuery)

	return searchEmailLog(c, elasticQuery)
}
