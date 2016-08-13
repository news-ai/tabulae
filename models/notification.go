package models

type Notification struct {
	Base
}

type NotificationObject struct {
	Base

	NoticationId int64

	Object   string
	ObjectId int64
}

type NotificationChange struct {
	Base

	NoticationObjectId int64

	Verb  string
	Actor string

	Read bool
}
