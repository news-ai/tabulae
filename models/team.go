package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Team struct {
	Base

	Members []int64 `json:"members"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (t *Team) key(c appengine.Context) *datastore.Key {
	if t.Id == 0 {
		return datastore.NewIncompleteKey(c, "Team", nil)
	}
	return datastore.NewKey(c, "Team", "", t.Id, nil)
}

/*
* Create methods
 */

// Function to create a new team into App Engine
func (t *Team) Create(c appengine.Context, r *http.Request, currentUser User) (*Team, error) {
	t.Members = append(t.Members, currentUser.Id)
	t.CreatedBy = currentUser.Id
	t.Created = time.Now()

	_, err := t.Save(c)
	return t, err
}

/*
* Update methods
 */

// Function to save a new team into App Engine
func (t *Team) Save(c appengine.Context) (*Team, error) {
	// Update the Updated time
	t.Updated = time.Now()

	// Save the object
	k, err := datastore.Put(c, t.key(c), t)
	if err != nil {
		return nil, err
	}
	t.Id = k.IntID()
	return t, nil
}
