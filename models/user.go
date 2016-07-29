package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/news-ai/tabulae/utils"

	"github.com/gorilla/context"
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
* Get methods
 */

func getUser(c appengine.Context, id int64) (User, error) {
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

func (u *User) create(c appengine.Context, r *http.Request) (*User, error) {
	// Create user
	u.IsAdmin = false
	u.Created = time.Now()
	_, err := u.save(c)
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

func (u *User) update(c appengine.Context, r *http.Request) (*User, error) {
	if len(u.Employers) == 0 {
		CreateAgencyFromUser(c, r, u)
	}
	return u, nil
}

/*
* Filter methods
 */

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

func GetUser(c appengine.Context, r *http.Request, id string) (User, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := GetCurrentUser(c, r)
		if err != nil {
			return User{}, err
		}
		return user, err
	default:
		userId, err := StringIdToInt(id)
		if err != nil {
			return User{}, err
		}
		user, err := getUser(c, userId)
		if err != nil {
			return User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}

func GetUserByEmail(c appengine.Context, email string) (User, error) {
	// Get the current user
	user, err := filterUser(c, "Email", email)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func GetUserByConfirmationCode(c appengine.Context, confirmationCode string) (User, error) {
	// Get the current user
	user, err := filterUser(c, "ConfirmationCode", confirmationCode)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func GetCurrentUser(c appengine.Context, r *http.Request) (User, error) {
	// Get the current user
	_, ok := context.GetOk(r, "user")
	if !ok {
		return User{}, errors.New("No user logged in")
	}
	user := context.Get(r, "user").(User)
	return user, nil
}

/*
* Create methods
 */

func RegisterUser(r *http.Request, user User) (bool, error) {
	c := appengine.NewContext(r)
	_, err := GetUserByEmail(c, user.Email)
	if err != nil {
		_, err = user.create(c, r)
		return true, nil
	}
	return false, errors.New("User with the email already exists")
}

func ValidateUserPassword(r *http.Request, email string, password string) (bool, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByEmail(c, email)
	if err == nil {
		err = utils.ValidatePassword(user.Password, password)
		if err != nil {
			return false, nil
		}
		return true, nil
	}
	return false, errors.New("User does not exist")
}

func NewOrUpdateUser(c appengine.Context, r *http.Request, email string, userDetails map[string]string) {
	_, ok := context.GetOk(r, "user")
	if !ok {
		user, err := GetUserByEmail(c, email)
		if err != nil {
			// Add the user if there is no user
			// If the registration comes from Google
			user := User{}
			user.GoogleId = userDetails["id"]
			user.Email = userDetails["email"]
			user.FirstName = userDetails["given_name"]
			user.LastName = userDetails["family_name"]
			user.EmailConfirmed = true
			_, err = user.create(c, r)
		} else {
			context.Set(r, "user", user)
			user.update(c, r)
		}
	} else {
		user := context.Get(r, "user").(User)
		user.update(c, r)
	}
}

/*
* Update methods
 */

func (u *User) ConfirmEmail(c appengine.Context) (*User, error) {
	u.EmailConfirmed = true
	u.ConfirmationCode = ""
	_, err := u.save(c)
	if err != nil {
		return u, err
	}
	return u, nil
}

func UpdateUser(c appengine.Context, r *http.Request, id string) (User, error) {
	// Get the details of the current user
	user, err := GetUser(c, r, id)
	if err != nil {
		return User{}, err
	}

	decoder := json.NewDecoder(r.Body)
	var updatedUser User
	err = decoder.Decode(&updatedUser)
	if err != nil {
		return User{}, err
	}

	UpdateIfNotBlank(&user.FirstName, updatedUser.FirstName)
	UpdateIfNotBlank(&user.LastName, updatedUser.LastName)

	if len(updatedUser.Employers) > 0 {
		user.Employers = updatedUser.Employers
	}

	user.save(c)
	return user, nil
}
