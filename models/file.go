package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type File struct {
	Base

	FileName string `json:"filename"`
	ListId   int64  `json:"listid"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (f *File) key(c appengine.Context) *datastore.Key {
	if f.Id == 0 {
		f.Created = time.Now()
		return datastore.NewIncompleteKey(c, "File", nil)
	}
	return datastore.NewKey(c, "File", "", f.Id, nil)
}

/*
* Create methods
 */

func (f *File) Create(c appengine.Context, r *http.Request, currentUser User) (*File, error) {
	f.CreatedBy = currentUser.Id
	f.Created = time.Now()

	_, err := f.Save(c)
	return f, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (f *File) Save(c appengine.Context) (*File, error) {
	// Update the Updated time
	f.Updated = time.Now()

	k, err := datastore.Put(c, f.key(c), f)
	if err != nil {
		return nil, err
	}
	f.Id = k.IntID()
	return f, nil
}
