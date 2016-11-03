package search

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticList *elastic.Elastic
)

type Lists struct {
	Type string `json:"type"`

	Archived   bool      `json:"archived"`
	Subscribed bool      `json:"subscribed"`
	PublicList bool      `json:"publiclist"`
	FileUpload int64     `json:"fileupload"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	CreatedBy  int64     `json:"createdby"`
	Client     string    `json:"client"`
	Name       string    `json:"name"`
	Id         int64     `json:"id"`
}

func (l *Lists) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(l, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchList(c context.Context, elasticQuery elastic.ElasticQueryMustShould) ([]Lists, error) {
	hits, err := elasticList.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Lists{}, err
	}

	listHits := hits.Hits
	lists := []Lists{}
	for i := 0; i < len(listHits); i++ {
		rawList := listHits[i].Source.Data
		rawMap := rawList.(map[string]interface{})
		contact := Lists{}
		err := contact.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		contact.Type = "lists"
		lists = append(lists, contact)
	}

	return lists, nil
}

func SearchListsByClientName(c context.Context, r *http.Request, clientName string, userId int64) ([]Lists, error) {
	if clientName == "" {
		return []Lists{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryMustShould{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticClientQuery := ElasticClientQuery{}
	elasticClientQuery.Term.Client = clientName
	elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticClientQuery)

	elasticQuery.MinScore = float32(0.5)

	return searchList(c, elasticQuery)
}
