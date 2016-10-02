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
	elasticContact *elastic.Elastic
)

func searchContact(c context.Context, elasticQuery elastic.ElasticQuery) ([]models.Contact, error) {
	hits, err := elasticContact.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, err
	}

	contactHits := hits.Hits
	contacts := []models.Contact{}
	for i := 0; i < len(contactHits); i++ {
		rawContact := contactHits[i].Source.Data
		rawMap := rawContact.(map[string]interface{})
		contact := models.Contact{}
		err := contact.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		contact.Type = "contacts"
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func SearchContacts(c context.Context, r *http.Request, search string, userId int64) ([]models.Contact, error) {
	if userId == 0 || search == "" {
		return []models.Contact{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = search

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	return searchContact(c, elasticQuery)
}

func SearchContactsByList(c context.Context, r *http.Request, search string, user models.User, listId int64) ([]models.Contact, error) {
	if listId == 0 || search == "" {
		return []models.Contact{}, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	if !user.IsAdmin {
		elasticCreatedByQuery := ElasticCreatedByQuery{}
		elasticCreatedByQuery.Term.CreatedBy = user.Id
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	}

	elasticListIdQuery := ElasticListIdQuery{}
	elasticListIdQuery.Term.ListId = listId

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = search

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticListIdQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	return searchContact(c, elasticQuery)
}
