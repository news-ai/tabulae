package models

import (
	"time"
)

type File struct {
	Id int64 `json:"id" datastore:"-"`

	FileName string `json:"filename"`
	ListId   int64  `json:"listid"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
