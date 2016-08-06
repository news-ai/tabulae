package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Agency struct {
	Id int64 `json:"id" datastore:"-"`

	Name  string `json:"name"`
	Email string `json:"email"`

	Administrators []int64 `json:"administrators"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (a *Agency) key(c appengine.Context) *datastore.Key {
	if a.Id == 0 {
		return datastore.NewIncompleteKey(c, "Agency", nil)
	}
	return datastore.NewKey(c, "Agency", "", a.Id, nil)
}

/*
* Create methods
 */

func (a *Agency) Create(c appengine.Context, r *http.Request, currentUser User) (*Agency, error) {
	a.CreatedBy = currentUser.Id
	a.Created = time.Now()
	_, err := a.save(c)
	return a, err
}

/*
* Update methods
 */

// Function to save a new agency into App Engine
func (a *Agency) save(c appengine.Context) (*Agency, error) {
	// Update the Updated time
	a.Updated = time.Now()

	k, err := datastore.Put(c, a.key(c), a)
	if err != nil {
		return nil, err
	}
	a.Id = k.IntID()
	return a, nil
}
