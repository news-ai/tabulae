package models

import (
	"encoding/json"
	"errors"
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

func GetUsers(c appengine.Context, id string) ([]User, error) {
	// Get the current signed in user details by Email
	users := []User{}
	ks, err := datastore.NewQuery("User").GetAll(c, &users)
	if err != nil {
		return []User{}, err
	}
	for i := 0; i < len(users); i++ {
		users[i].Id = ks[i].IntID()
	}
	return users, nil
}

func getUser(c appengine.Context, id string) (User, error) {
	// Get the current signed in user details by Email
	users := []User{}
	ks, err := datastore.NewQuery("User").Filter("Id =", id).GetAll(c, &users)
	if err != nil {
		return User{}, err
	}
	if len(users) > 0 {
		users[0].Id = ks[0].IntID()
		return users[0], nil
	}
	return User{}, errors.New("No user by this email")
}

func getCurrentUser(c appengine.Context) (User, error) {
	// Get the current signed in user details by Email
	users := []User{}
	currentUser := user.Current(c)
	ks, err := datastore.NewQuery("User").Filter("Email =", currentUser.Email).GetAll(c, &users)
	if err != nil {
		return User{}, err
	}
	if len(users) > 0 {
		users[0].Id = ks[0].IntID()
		return users[0], nil
	}
	return User{}, errors.New("No user by this email")
}

func createUser(c appengine.Context) User {
	currentUser := user.Current(c)
	newUser := User{}
	newUser.Email = currentUser.ID
	newUser.GoogleId = currentUser.Email
	newUser.save(c)
	return newUser
}

func GetUser(c appengine.Context, id string) (User, error) {
	// Get the details of the current user
	if id == "me" {
		user, err := getCurrentUser(c)

		if err != nil {
			// Add the user if there is no user
			newUser := createUser(c)
			user = newUser
		}
		return user, nil
	} else {
		user, err := getUser(c, id)
		if err != nil {
			return User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}
