package models

import (
	"time"
)

type List struct {
	Id int64 `json:"id" datastore:"-"`

	Contacts []int64 `json:"contacts"`

	CreatedBy int64 `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
}
