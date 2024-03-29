package search

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	"github.com/news-ai/api/models"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticAgency *elastic.Elastic
)

func SearchAgency(c context.Context, r *http.Request, search string) ([]models.Agency, int, error) {
	search = url.QueryEscape(search)
	search = "q=data.Name:" + search

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	hits, err := elasticAgency.Query(c, offset, limit, search)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Agency{}, 0, err
	}

	agencyHits := hits.Hits
	agencies := []models.Agency{}
	for i := 0; i < len(agencyHits); i++ {
		rawAgency := agencyHits[i].Source.Data
		rawMap := rawAgency.(map[string]interface{})
		agency := models.Agency{}
		err := agency.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		agency.Type = "agencies"
		agencies = append(agencies, agency)
	}

	return agencies, hits.Total, nil
}
