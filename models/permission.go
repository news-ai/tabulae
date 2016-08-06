package models

type ListPermission struct {
	Base

	ListId int64 `json:"listid"`
	UserId int64 `json:"userid"`

	CanWrite bool `json:"permissionlevel"`
}
