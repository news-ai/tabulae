package models

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type Base struct {
	Id int64 `json:"id" datastore:"-"`

	Type string `json:"type" datastore:"-"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (b *Base) key(c context.Context, collection string) *datastore.Key {
	if b.Id == 0 {
		return datastore.NewIncompleteKey(c, collection, nil)
	}
	return datastore.NewKey(c, collection, "", b.Id, nil)
}
