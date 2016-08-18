package models

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type CustomContactField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BaseResponse struct {
	Count    int         `json:"count"`
	Next     string      `json:"-"`
	Results  interface{} `json:"results"`
	Includes interface{} `json:"includes"`
}

type Base struct {
	Id int64 `json:"id" datastore:"-"`

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
