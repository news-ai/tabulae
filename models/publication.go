package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Publication struct {
	Id int64 `json:"id" datastore:"-"`

	Name string `json:"name"`
	Url  string `json:"url"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (p *Publication) key(c appengine.Context) *datastore.Key {
	if p.Id == 0 {
		return datastore.NewIncompleteKey(c, "Publication", nil)
	}
	return datastore.NewKey(c, "Publication", "", p.Id, nil)
}

/*
* Create methods
 */

// Function to create a new publication into App Engine
func (p *Publication) Create(c appengine.Context, r *http.Request, currentUser User) (*Publication, error) {
	p.CreatedBy = currentUser.Id
	p.Created = time.Now()

	_, err := p.Save(c)
	return p, err
}

/*
* Update methods
 */

// Function to save a new publication into App Engine
func (p *Publication) Save(c appengine.Context) (*Publication, error) {
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
