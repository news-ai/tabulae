package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
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

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (u *User) key(c appengine.Context) *datastore.Key {
	if u.Id == 0 {
		return datastore.NewIncompleteKey(c, "User", nil)
	}
	return datastore.NewKey(c, "User", "", u.Id, nil)
}

/*
* Create methods
 */

func (u *User) Create(c appengine.Context, r *http.Request) (*User, error) {
	// Create user
	u.IsAdmin = false
	u.Created = time.Now()
	_, err := u.Save(c)
	return u, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (u *User) Save(c appengine.Context) (*User, error) {
	u.Updated = time.Now()

	k, err := datastore.Put(c, u.key(c), u)
	if err != nil {
		return nil, err
	}
	u.Id = k.IntID()
	return u, nil
}

func (u *User) ConfirmEmail(c appengine.Context) (*User, error) {
	u.EmailConfirmed = true
	u.ConfirmationCode = ""
	_, err := u.Save(c)
	if err != nil {
		return u, err
	}
	return u, nil
}
