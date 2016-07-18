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

	WorksAt []Agency `json:"worksat"`

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
* Get methods
 */

func getUser(c appengine.Context, queryType, query string) (User, error) {
	// Get the current signed in user details by Id
	users := []User{}
	ks, err := datastore.NewQuery("User").Filter(queryType+" =", query).GetAll(c, &users)
	if err != nil {
		return User{}, err
	}
	if len(users) > 0 {
		users[0].Id = ks[0].IntID()
		return users[0], nil
	}
	return User{}, errors.New("No user by this " + queryType)
}

/*
* Create methods
 */

func (u *User) create(c appengine.Context) (*User, error) {
	// Create user
	currentUser := user.Current(c)
	u.Email = currentUser.Email
	u.GoogleId = currentUser.ID
	u.Created = time.Now()
	_, err := u.save(c)
	CreateAgencyFromUser(c, u)
	return u, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (u *User) save(c appengine.Context) (*User, error) {
	u.Updated = time.Now()
	k, err := datastore.Put(c, u.key(c), u)
	if err != nil {
		return nil, err
	}
	u.Id = k.IntID()
	return u, nil
}

func (u *User) update(c appengine.Context) (*User, error) {
	if len(u.WorksAt) == 0 {
		CreateAgencyFromUser(c, u)
	}
	return u, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

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

func GetUser(c appengine.Context, id string) (User, error) {
	// Get the details of the current user
	switch id {
	case "me":
		currentUser := user.Current(c)
		user, err := getUser(c, "Email", currentUser.Email)
		return user, err
	default:
		user, err := getUser(c, "Id", id)
		if err != nil {
			return User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}

func GetCurrentUser(c appengine.Context) (User, error) {
	// Get the current user
	currentUser := user.Current(c)
	user, err := getUser(c, "Email", currentUser.Email)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

/*
* Create methods
 */

func NewOrUpdateUser(c appengine.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		// Add the user if there is no user
		user := User{}
		_, err = user.create(c)
	} else {
		user.update(c)
	}
}
