package models

import (
	apiModels "github.com/news-ai/api/models"
)

type ListPermission struct {
	apiModels.Base

	ListId int64 `json:"listid" apiModel:"List"`
	UserId int64 `json:"userid" apiModel:"User"`

	CanWrite bool `json:"permissionlevel"`
}
