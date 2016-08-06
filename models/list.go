package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type MediaList struct {
	Base

	Name string `json:"name"`

	Contacts     []int64  `json:"contacts"`
	CustomFields []string `json:"customfields"`

	FileUpload int64 `json:"fileupload"`

	Archived bool `json:"archived"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (ml *MediaList) key(c appengine.Context) *datastore.Key {
	if ml.Id == 0 {
		ml.Created = time.Now()
		return datastore.NewIncompleteKey(c, "MediaList", nil)
	}
	return datastore.NewKey(c, "MediaList", "", ml.Id, nil)
}

/*
* Create methods
 */

func (ml *MediaList) Create(c appengine.Context, r *http.Request, currentUser User) (*MediaList, error) {
	ml.CreatedBy = currentUser.Id
	ml.Created = time.Now()

	_, err := ml.Save(c)
	return ml, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ml *MediaList) Save(c appengine.Context) (*MediaList, error) {
	// Update the Updated time
	ml.Updated = time.Now()

	k, err := datastore.Put(c, ml.key(c), ml)
	if err != nil {
		return nil, err
	}
	ml.Id = k.IntID()
	return ml, nil
}
