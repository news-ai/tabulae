package search

import (
	"net/http"
	"time"
	// "net/url"
	// "strconv"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticHeadline *elastic.Elastic
)

type Headline struct {
	Type string `json:"type"`

	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Url         string    `json:"url"`
	Categories  []string  `json:"categories"`
	PublishDate time.Time `json:"publishdate"`
	Summary     string    `json:"summary"`

	ContactId int64 `json:"contactid"`
	ListId    int64 `json:"listid"`
}

func (h *Headline) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(h, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchHeadline(c context.Context, elasticQuery elastic.ElasticQuery) ([]Headline, error) {
	hits, err := elasticHeadline.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Headline{}, err
	}

	headlineHits := hits.Hits
	headlines := []Headline{}
	for i := 0; i < len(headlineHits); i++ {
		rawHeadline := headlineHits[i].Source.Data
		rawMap := rawHeadline.(map[string]interface{})
		headline := Headline{}
		err := headline.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		headline.Type = "headlines"
		headlines = append(headlines, headline)
	}

	return headlines, nil
}

func SearchHeadlines(c context.Context, r *http.Request, contactId int64) ([]Headline, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticContactIdQuery := ElasticContactIdQuery{}
	elasticContactIdQuery.Term.ContactId = contactId

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticContactIdQuery)

	return searchHeadline(c, elasticQuery)
}
