package models

import (
	"time"
)

type Email struct {
	Id int64 `json:"id" datastore:"-"`

	// Which list it belongs to
	ListId int64 `json:"listid"`

	Sender  string `json:"sender"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	IsSent bool `json:"issent"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
