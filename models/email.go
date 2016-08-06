package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Email struct {
	Base

	// Which list it belongs to
	ListId int64 `json:"listid"`

	Sender  string `json:"sender"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	IsSent bool `json:"issent"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (e *Email) key(c appengine.Context) *datastore.Key {
	if e.Id == 0 {
		e.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Email", nil)
	}
	return datastore.NewKey(c, "Email", "", e.Id, nil)
}

/*
* Create methods
 */

func (e *Email) Create(c appengine.Context, r *http.Request, currentUser User) (*Email, error) {
	e.IsSent = false
	e.CreatedBy = currentUser.Id
	e.Created = time.Now()

	_, err := e.Save(c)
	return e, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (e *Email) Save(c appengine.Context) (*Email, error) {
	// Update the Updated time
	e.Updated = time.Now()

	k, err := datastore.Put(c, e.key(c), e)
	if err != nil {
		return nil, err
	}
	e.Id = k.IntID()
	return e, nil
}

func (e *Email) MarkSent(c appengine.Context) (*Email, error) {
	e.IsSent = true
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}
