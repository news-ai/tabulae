package models

import (
	"time"
)

type ListPermission struct {
	Id int64 `json:"id" datastore:"-"`

	ListId int64 `json:"listid"`
	UserId int64 `json:"userid"`

	CanWrite bool `json:"permissionlevel"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
