package models

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type Publication struct {
	Id int64 `json:"id" datastore:"-"`

	Name string `json:"name"`
	Url  string `json:"url"`

	CreatedBy string `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Code to get data from App Engine
func defaultPublicationList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "PublicationList", "default", 0, nil)
}

// Generates a new key for the data to be stored on App Engine
func (p *Publication) key(c appengine.Context) *datastore.Key {
	if p.Id == 0 {
		p.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Publication", defaultPublicationList(c))
	}
	return datastore.NewKey(c, "Publication", "", p.Id, defaultPublicationList(c))
}

// Function to save a new publication into App Engine
func (p *Publication) save(c appengine.Context) (*Publication, error) {
	k, err := datastore.Put(c, p.key(c), p)
	if err != nil {
		return nil, err
	}
	p.Id = k.IntID()
	return p, nil
}

// Function to create a new publication into App Engine
func (p *Publication) create(c appengine.Context) (*Publication, error) {
	currentUser := user.Current(c)

	p.CreatedBy = currentUser.ID
	p.Created = time.Now()

	_, err := p.save(c)
	return p, err
}

func getPublication(c appengine.Context, id string) (Publication, error) {
	// Get the publication details by id
	publications := []Publication{}
	ks, err := datastore.NewQuery("Publication").Filter("ID =", id).GetAll(c, &publications)
	if err != nil {
		return Publication{}, err
	}
	if len(publications) > 0 {
		publications[0].Id = ks[0].IntID()
		return publications[0], nil
	}
	return Publication{}, errors.New("No publication by this id")
}

func getPublicationByUrl(c appengine.Context, url string) (Publication, error) {
	// Get a publication by the URL
	publications := []Publication{}
	ks, err := datastore.NewQuery("Publication").Filter("Url =", url).GetAll(c, &publications)
	if err != nil {
		return Publication{}, err
	}
	if len(publications) > 0 {
		publications[0].Id = ks[0].IntID()
		return publications[0], nil
	}
	return Publication{}, errors.New("No publication by this url")
}

/*
* Public methods
 */

func GetPublicationByEmail(c appengine.Context, email string) (Publication, error) {
	// Get the id of the current publication
	publication, err := getPublicationByUrl(c, email)
	if err != nil {
		return Publication{}, err
	}
	return publication, nil
}

func GetPublications(c appengine.Context) ([]Publication, error) {
	publications := []Publication{}
	ks, err := datastore.NewQuery("Publication").GetAll(c, &publications)
	if err != nil {
		return []Publication{}, err
	}
	for i := 0; i < len(publications); i++ {
		publications[i].Id = ks[i].IntID()
	}
	return publications, nil
}

func GetPublication(c appengine.Context, id string) (Publication, error) {
	// Get a publication by id
	publication, err := getPublication(c, id)
	if err != nil {
		return Publication{}, err
	}
	return publication, nil
}
