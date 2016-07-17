package models

import (
	"time"
)

type Publication struct {
	Id int64 `json:"id" datastore:"-"`

	Name string `json:"name"`
	Url  string `json:"url"`

	CreatedBy int64 `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
}
