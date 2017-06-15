package search

import (
	"net/http"
	// "net/url"
	// "strconv"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"
	apiSearch "github.com/news-ai/api/search"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticContact *elastic.Elastic
)

func searchContact(c context.Context, elasticQuery elastic.ElasticQuery) ([]models.Contact, int, error) {
	hits, err := elasticContact.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, 0, err
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

	return contacts, hits.Total, nil
}

func SearchContacts(c context.Context, r *http.Request, search string, userId int64) ([]models.Contact, int, error) {
	if userId == 0 || search == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = search

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	return searchContact(c, elasticQuery)
}

func SearchContactsByList(c context.Context, r *http.Request, search string, user apiModels.User, userId int64, listId int64) ([]models.Contact, int, error) {
	if listId == 0 || search == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	if !user.IsAdmin {
		elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
		elasticCreatedByQuery.Term.CreatedBy = userId
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	}

	elasticListIdQuery := apiSearch.ElasticListIdQuery{}
	elasticListIdQuery.Term.ListId = listId

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = search

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticListIdQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	return searchContact(c, elasticQuery)
}

func SearchContactsByTag(c context.Context, r *http.Request, tag string, userId int64) ([]models.Contact, int, error) {
	if tag == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticTagQuery := apiSearch.ElasticTagQuery{}
	elasticTagQuery.Term.Tag = tag
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticTagQuery)

	return searchContact(c, elasticQuery)
}

func SearchContactsByPublicationId(c context.Context, r *http.Request, publicationId string, userId int64) ([]models.Contact, int, error) {
	if publicationId == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticEmployersQuery := apiSearch.ElasticEmployersQuery{}
	elasticEmployersQuery.Term.Employers = publicationId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmployersQuery)

	return searchContact(c, elasticQuery)
}

func SearchContactsByFieldSelector(c context.Context, r *http.Request, fieldSelector string, query string, userId int64) ([]models.Contact, int, error) {
	if fieldSelector == "tag" {
		return SearchContactsByTag(c, r, query, userId)
	} else if fieldSelector == "publication" {
		return SearchContactsByPublicationId(c, r, query, userId)
	}

	return []models.Contact{}, 0, nil
}
