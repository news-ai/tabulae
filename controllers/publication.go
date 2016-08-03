package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/news-ai/tabulae/models"
)

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (p *models.Publication) key(c appengine.Context) *datastore.Key {
	if p.Id == 0 {
		return datastore.NewIncompleteKey(c, "Publication", nil)
	}
	return datastore.NewKey(c, "Publication", "", p.Id, nil)
}

/*
* Get methods
 */

func getPublication(c appengine.Context, id int64) (models.Publication, error) {
	// Get the publication details by id
	publications := []models.Publication{}
	publicationId := datastore.NewKey(c, "Publication", "", id, nil)
	ks, err := datastore.NewQuery("Publication").Filter("__key__ =", publicationId).GetAll(c, &publications)
	if err != nil {
		return models.Publication{}, err
	}

	if len(publications) > 0 {
		publications[0].Id = ks[0].IntID()
		return publications[0], nil
	}
	return models.Publication{}, errors.New("No publication by this id")
}

/*
* Create methods
 */

// Function to create a new publication into App Engine
func (p *models.Publication) create(c appengine.Context, r *http.Request) (*models.Publication, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return p, err
	}

	p.CreatedBy = currentUser.Id
	p.Created = time.Now()

	_, err = p.save(c)
	return p, err
}

/*
* Update methods
 */

// Function to save a new publication into App Engine
func (p *models.Publication) save(c appengine.Context) (*models.Publication, error) {
	// Update the Updated time
	p.Updated = time.Now()

	// Save the object
	k, err := datastore.Put(c, p.key(c), p)
	if err != nil {
		return nil, err
	}
	p.Id = k.IntID()
	return p, nil
}

/*
* Filter methods
 */

func filterPublication(c appengine.Context, queryType, query string) (models.Publication, error) {
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

func GetPublications(c appengine.Context) ([]models.Publication, error) {
	publications := []models.Publication{}
	ks, err := datastore.NewQuery("Publication").GetAll(c, &publications)
	if err != nil {
		return []models.Publication{}, err
	}

	for i := 0; i < len(publications); i++ {
		publications[i].Id = ks[i].IntID()
	}
	return publications, nil
}

func GetPublication(c appengine.Context, id string) (models.Publication, error) {
	// Get a publication by id
	currentId, err := StringIdToInt(id)
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

func CreatePublication(c appengine.Context, w http.ResponseWriter, r *http.Request) (models.Publication, error) {
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
		publication.Url, err = NormalizeUrl(publication.Url)
		if err != nil {
			return models.Publication{}, err
		}
	}

	presentPublication, err := FilterPublicationByName(c, publication.Name)
	if err != nil {
		// Create publication
		_, err = publication.create(c, r)
		if err != nil {
			return models.Publication{}, err
		}
		return publication, nil
	}

	return presentPublication, nil
}

func FindOrCreatePublication(c appengine.Context, r *http.Request, name string) (models.Publication, error) {
	name = strings.Trim(name, " ")
	publication, err := FilterPublicationByName(c, name)
	if err != nil {
		var newPublication models.Publication
		newPublication.Name = name
		_, err = newPublication.create(c, r)
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

func FilterPublicationByUrl(c appengine.Context, url string) (models.Publication, error) {
	// Get the id of the current publication
	publication, err := filterPublication(c, "Url", url)
	if err != nil {
		return models.Publication{}, err
	}
	return publication, nil
}

func FilterPublicationByName(c appengine.Context, name string) (models.Publication, error) {
	// Get the id of the current publication
	publication, err := filterPublication(c, "Name", name)
	if err != nil {
		return models.Publication{}, err
	}
	return publication, nil
}
