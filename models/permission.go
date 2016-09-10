package models

type ListPermission struct {
	Base

	ListId int64 `json:"listid" apiModel:"List"`
	UserId int64 `json:"userid" apiModel:"User"`

	CanWrite bool `json:"permissionlevel"`
}
