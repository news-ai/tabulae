package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	gcontext "github.com/gorilla/context"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getPublication(c context.Context, id int64) (models.Publication, error) {
	if id == 0 {
		return models.Publication{}, errors.New("datastore: no such entity")
	}
	// Get the publication details by id
	var publication models.Publication
	publicationId := datastore.NewKey(c, "Publication", "", id, nil)

	err := nds.Get(c, publicationId, &publication)

	if err != nil {
		return models.Publication{}, err
	}

	if !publication.Created.IsZero() {
		publication.Id = publicationId.IntID()
		return publication, nil
	}
	return models.Publication{}, errors.New("No publication by this id")
}

/*
* Filter methods
 */

func filterPublication(c context.Context, queryType, query string) (models.Publication, error) {
	// Get a publication by the URL
	publications := []models.Publication{}
	ks, err := datastore.NewQuery("Publication").Filter(queryType+" =", query).GetAll(c, &publications)
	if err != nil {
		return models.Publication{}, err
	}
	if len(publications) > 0 {
		publications[0].Id = ks[0].IntID()
		return publications[0], nil
	}
	return models.Publication{}, errors.New("No publication by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetPublications(c context.Context, r *http.Request) ([]models.Publication, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Publication{}, err
	}

	if !user.IsAdmin {
		return []models.Publication{}, errors.New("Forbidden")
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	publications := []models.Publication{}
	ks, err := datastore.NewQuery("Publication").Limit(limit).Offset(offset).GetAll(c, &publications)
	if err != nil {
		return []models.Publication{}, err
	}

	for i := 0; i < len(publications); i++ {
		publications[i].Id = ks[i].IntID()
	}
	return publications, nil
}

func GetPublication(c context.Context, id string) (models.Publication, error) {
	// Get a publication by id
	currentId, err := utils.StringIdToInt(id)
	if err != nil {
		return models.Publication{}, err
	}

	publication, err := getPublication(c, currentId)
	if err != nil {
		return models.Publication{}, err
	}
	return publication, nil
}

/*
* Create methods
 */

func CreatePublication(c context.Context, w http.ResponseWriter, r *http.Request) (models.Publication, error) {
	// Parse JSON
	decoder := json.NewDecoder(r.Body)
	var publication models.Publication
	err := decoder.Decode(&publication)
	if err != nil {
		return models.Publication{}, err
	}

	// Validate Fields
	if publication.Name == "" {
		return models.Publication{}, errors.New("Missing fields")
	}

	// Format URL properly
	if publication.Url != "" {
		publication.Url, err = utils.NormalizeUrl(publication.Url)
		if err != nil {
			return models.Publication{}, err
		}
	}

	presentPublication, err := FilterPublicationByName(c, publication.Name)
	if err != nil {
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			return models.Publication{}, err
		}

		// Create publication
		_, err = publication.Create(c, r, currentUser)
		if err != nil {
			return models.Publication{}, err
		}
		return publication, nil
	}

	return presentPublication, nil
}

func FindOrCreatePublication(c context.Context, r *http.Request, name string) (models.Publication, error) {
	name = strings.Trim(name, " ")
	publication, err := FilterPublicationByName(c, name)
	if err != nil {
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			return models.Publication{}, err
		}

		var newPublication models.Publication
		newPublication.Name = name
		_, err = newPublication.Create(c, r, currentUser)
		if err != nil {
			return models.Publication{}, err
		}
		return newPublication, nil
	}

	return publication, nil
}

/*
* Filter methods
 */

func FilterPublicationByUrl(c context.Context, url string) (models.Publication, error) {
	// Get the id of the current publication
	publication, err := filterPublication(c, "Url", url)
	if err != nil {
		return models.Publication{}, err
	}
	return publication, nil
}

func FilterPublicationByName(c context.Context, name string) (models.Publication, error) {
	// Get the id of the current publication
	publication, err := filterPublication(c, "Name", name)
	if err != nil {
		return models.Publication{}, err
	}
	return publication, nil
}
