package models

import (
	"errors"
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

// Code to get data from App Engine
func defaultUserList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "UserList", "default", 0, nil)
}

// Generates a new key for the data to be stored on App Engine
func (u *User) key(c appengine.Context) *datastore.Key {
	if u.Id == 0 {
		u.Created = time.Now()
		return datastore.NewIncompleteKey(c, "User", defaultUserList(c))
	}
	return datastore.NewKey(c, "User", "", u.Id, defaultUserList(c))
}

// Function to save a new user into App Engine
func (u *User) save(c appengine.Context) (*User, error) {
	k, err := datastore.Put(c, u.key(c), u)
	if err != nil {
		return nil, err
	}
	u.Id = k.IntID()
	return u, nil
}

func (u *User) create(c appengine.Context) (*User, error) {
	currentUser := user.Current(c)
	u.Email = currentUser.ID
	u.GoogleId = currentUser.Email
	_, err := u.save(c)
	return u, err
}

// Gets every single user
func GetUsers(c appengine.Context) ([]User, error) {
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
	ks, err := datastore.NewQuery("User").Filter("ID =", id).GetAll(c, &users)
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

func GetUser(c appengine.Context, id string) (User, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := getCurrentUser(c)
		if err != nil {
			// Add the user if there is no user
			user := User{}
			_, err = user.create(c)
		}
		return user, nil
	default:
		user, err := getUser(c, id)
		if err != nil {
			return User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}
