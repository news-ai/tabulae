package models

import (
	"time"
)

type MediaList struct {
	Id int64 `json:"id" datastore:"-"`

	Name string `json:"name"`

	Contacts     []int64  `json:"contacts"`
	CustomFields []string `json:"customfields"`

	FileUpload int64 `json:"fileupload"`

	Archived bool `json:"archived"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
