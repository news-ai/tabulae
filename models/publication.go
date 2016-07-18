package models

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type Publication struct {
	Id int64 `json:"id" datastore:"-"`

	Name string `json:"name"`
	Url  string `json:"url"`

	CreatedBy User `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (p *Publication) key(c appengine.Context) *datastore.Key {
	if p.Id == 0 {
		p.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Publication", nil)
	}
	return datastore.NewKey(c, "Publication", "", 0, nil)
}

/*
* Get methods
 */

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

func getPublicationByName(c appengine.Context, name string) (Publication, error) {
	// Get a publication by the URL
	publications := []Publication{}
	ks, err := datastore.NewQuery("Publication").Filter("Name =", name).GetAll(c, &publications)
	if err != nil {
		return Publication{}, err
	}
	if len(publications) > 0 {
		publications[0].Id = ks[0].IntID()
		return publications[0], nil
	}
	return Publication{}, errors.New("No publication by this name")
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
* Create methods
 */

// Function to create a new publication into App Engine
func (p *Publication) create(c appengine.Context) (*Publication, error) {
	currentUser, err := GetCurrentUser(c)
	if err != nil {
		return p, err
	}

	p.CreatedBy = currentUser
	p.Created = time.Now()

	_, err = p.save(c)
	return p, err
}

/*
* Update methods
 */

// Function to save a new publication into App Engine
func (p *Publication) save(c appengine.Context) (*Publication, error) {
	k, err := datastore.Put(c, p.key(c), p)
	if err != nil {
		return nil, err
	}
	p.Id = k.IntID()
	return p, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

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

func GetPublicationByUrl(c appengine.Context, url string) (Publication, error) {
	// Get the id of the current publication
	publication, err := getPublicationByUrl(c, url)
	if err != nil {
		return Publication{}, err
	}
	return publication, nil
}

func GetPublicationByName(c appengine.Context, name string) (Publication, error) {
	// Get the id of the current publication
	publication, err := getPublicationByName(c, name)
	if err != nil {
		return Publication{}, err
	}
	return publication, nil
}

/*
* Create methods
 */

func CreatePublicationFromName(c appengine.Context, name string) (Publication, error) {
	publication := Publication{}
	publication.Name = name
	_, err := publication.create(c)
	if err != nil {
		return Publication{}, err
	}
	return publication, nil
}
