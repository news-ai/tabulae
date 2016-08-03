package models

import (
	"time"
)

type User struct {
	Id       int64  `json:"id" datastore:"-"`
	GoogleId string `json:"googleid"`

	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	Password []byte `json:"-"`

	Employers []int64 `json:"employers"`

	ConfirmationCode string `json:"-"`
	EmailConfirmed   bool   `json:"emailconfirmed"`
	IsAdmin          bool   `json:"-"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
