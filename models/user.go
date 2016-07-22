package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"

	"github.com/gorilla/context"
)

type User struct {
	Id       int64  `json:"id" datastore:"-"`
	GoogleId string `json:"googleid"`

	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	Employers []int64 `json:"employers"`

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

func getUserById(c appengine.Context, id int64) (User, error) {
	// Get the current signed in user details by Id
	users := []User{}
	userId := datastore.NewKey(c, "User", "", id, nil)
	ks, err := datastore.NewQuery("User").Filter("__key__ =", userId).GetAll(c, &users)
	if err != nil {
		return User{}, err
	}

	if len(users) > 0 {
		users[0].Id = ks[0].IntID()
		return users[0], nil
	}
	return User{}, errors.New("No user by this id")
}

func filterUser(c appengine.Context, queryType, query string) (User, error) {
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

// Gets every single user
func getUsers(c appengine.Context) ([]User, error) {
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
	if len(u.Employers) == 0 {
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

func GetUsers(c appengine.Context) ([]User, error) {
	// Get the current user
	users, err := getUsers(c)
	if err != nil {
		return []User{}, err
	}
	return users, nil
}

func GetUser(c appengine.Context, id string) (User, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := GetCurrentUser(c)
		if err != nil {
			return User{}, err
		}
		return user, err
	default:
		userId, err := StringIdToInt(id)
		if err != nil {
			return User{}, err
		}
		user, err := getUserById(c, userId)
		if err != nil {
			return User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}

func GetCurrentUser(c appengine.Context) (User, error) {
	// Get the current user
	currentUser := user.Current(c)
	user, err := filterUser(c, "Email", currentUser.Email)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

/*
* Create methods
 */

func NewOrUpdateUser(c appengine.Context, r *http.Request) {
	_, ok := context.GetOk(r, "user")
	if !ok {
		user, err := GetCurrentUser(c)
		if err != nil {
			// Add the user if there is no user
			user := User{}
			_, err = user.create(c)
		} else {
			user.update(c)
		}
		context.Set(r, "user", user)
	} else {
		user := context.Get(r, "user").(User)
		user.update(c)
	}
}

/*
* Update methods
 */

func UpdateUser(c appengine.Context, r *http.Request, id string) (User, error) {
	// Get the details of the current user
	user, err := GetUser(c, id)
	if err != nil {
		return User{}, err
	}

	decoder := json.NewDecoder(r.Body)
	var updatedUser User
	err = decoder.Decode(&updatedUser)
	if err != nil {
		return User{}, err
	}

	user.FirstName = updatedUser.FirstName
	user.LastName = updatedUser.LastName

	user.save(c)
	return user, nil
}
