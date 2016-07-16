package models

import (
	"encoding/json"
	"io"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type User struct {
	Id       int64  `json:"id" datastore:"-"`
	GoogleId string `json:"googleid"`

	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	Agency int64 `json:"agencyid" datastore:"-"`

	Created time.Time `json:"created"`
}

func defaultUserList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "UserList", "default", 0, nil)
}

func (u *User) key(c appengine.Context) *datastore.Key {
	if u.Id == 0 {
		u.Created = time.Now()
		return datastore.NewIncompleteKey(c, "User", defaultUserList(c))
	}
	return datastore.NewKey(c, "User", "", u.Id, defaultUserList(c))
}

func (u *User) save(c appengine.Context) (*User, error) {
	k, err := datastore.Put(c, u.key(c), u)
	if err != nil {
		return nil, err
	}
	u.Id = k.IntID()
	return u, nil
}

func decodeUser(r io.ReadCloser) (*User, error) {
	defer r.Close()
	var user User
	err := json.NewDecoder(r).Decode(&user)
	return &user, err
}

func createNewUser(c appengine.Context, googleId string, email string) User {
	newUser := User{}
	newUser.Email = email
	newUser.GoogleId = googleId
	newUser.save(c)
	return newUser
}

func GetCurrentUser(c appengine.Context) (User, error) {
	currentUser := user.Current(c)
	users := []User{}
	ks, err := datastore.NewQuery("User").Filter("Email =", currentUser.Email).GetAll(c, &users)
	if err != nil {
		return User{}, err
	}
	for i := 0; i < len(users); i++ {
		users[i].Id = ks[i].IntID()
	}

	// If there is a user
	if len(users) > 0 {
		return users[0], nil
	}

	// Add the user if there is no user
	newUser := createNewUser(c, currentUser.ID, currentUser.Email)
	return newUser, nil
}
