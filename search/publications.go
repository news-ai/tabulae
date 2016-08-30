package search

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticPublication *elastic.Elastic
)

func SearchPublication(c context.Context, r *http.Request, search string) ([]models.Publication, error) {
	search = url.QueryEscape(search)
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	hits, err := elasticPublication.Query(c, offset, limit, search)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Publication{}, err
	}

	publicationHits := hits.Hits
	publications := []models.Publication{}
	for i := 0; i < len(publicationHits); i++ {
		rawPublication := publicationHits[i].Source.Data
		rawMap := rawPublication.(map[string]interface{})
		publication := models.Publication{}
		err := publication.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		publication.Type = "publications"
		publications = append(publications, publication)
	}

	return publications, nil
}
