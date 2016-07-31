package models

import (
	"time"
)

type EmailUser struct {
	Id int64  `json:"id" datastore:"-"`
	To string `json:"to"`

	SeenAt []time.Time `json:"seenat"`
}

type Email struct {
	// Email details
	Sender  string
	To      []EmailUser
	Subject string
	Body    string

	// User details
	FirstName string
}
