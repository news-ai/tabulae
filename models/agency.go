package models

import (
	"time"

	"appengine"
	"appengine/datastore"
)

type Agency struct {
	Id int64 `json:"id" datastore:"-"`

	AgencyName string `json:"agencyname"`

	Created time.Time `json:"created"`
}

func defaultAgencyList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "AgencyList", "default", 0, nil)
}
