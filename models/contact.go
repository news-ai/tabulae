package models

import (
	"time"
)

type Contact struct {
	Id int64 `json:"id" datastore:"-"`

	Name  string `json:"name"`
	Email string `json:"email"`

	WorksAt []int64 `json:"worksat" datastore:"-"`

	CreatedBy int64 `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
}
