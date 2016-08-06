package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"appengine"
	"appengine/datastore"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"

	"github.com/gorilla/context"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getUser(c appengine.Context, id int64) (models.User, error) {
	// Get the current signed in user details by Id
	users := []models.User{}
	userId := datastore.NewKey(c, "User", "", id, nil)
	ks, err := datastore.NewQuery("User").Filter("__key__ =", userId).GetAll(c, &users)
	if err != nil {
		return models.User{}, err
	}

	if len(users) > 0 {
		users[0].Id = ks[0].IntID()
		return users[0], nil
	}
	return models.User{}, errors.New("No user by this id")
}

// Gets every single user
func getUsers(c appengine.Context) ([]models.User, error) {
	// Get the current signed in user details by Email
	users := []models.User{}
	ks, err := datastore.NewQuery("User").GetAll(c, &users)
	if err != nil {
		return []models.User{}, err
	}

	for i := 0; i < len(users); i++ {
		users[i].Id = ks[i].IntID()
	}
	return users, nil
}

/*
* Filter methods
 */

func filterUser(c appengine.Context, queryType, query string) (models.User, error) {
	// Get the current signed in user details by Id
	users := []models.User{}
	ks, err := datastore.NewQuery("User").Filter(queryType+" =", query).GetAll(c, &users)
	if err != nil {
		return models.User{}, err
	}

	if len(users) > 0 {
		users[0].Id = ks[0].IntID()
		return users[0], nil
	}
	return models.User{}, errors.New("No user by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetUsers(c appengine.Context) ([]models.User, error) {
	// Get the current user
	users, err := getUsers(c)
	if err != nil {
		return []models.User{}, err
	}
	return users, nil
}

func GetUser(c appengine.Context, r *http.Request, id string) (models.User, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.User{}, err
		}
		return user, err
	default:
		userId, err := StringIdToInt(id)
		if err != nil {
			return models.User{}, err
		}
		user, err := getUser(c, userId)
		if err != nil {
			return models.User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}

func GetUserByEmail(c appengine.Context, email string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "Email", email)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByConfirmationCode(c appengine.Context, confirmationCode string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ConfirmationCode", confirmationCode)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetCurrentUser(c appengine.Context, r *http.Request) (models.User, error) {
	// Get the current user
	_, ok := context.GetOk(r, "user")
	if !ok {
		return models.User{}, errors.New("No user logged in")
	}
	user := context.Get(r, "user").(models.User)
	return user, nil
}

/*
* Create methods
 */

func RegisterUser(r *http.Request, user models.User) (bool, error) {
	c := appengine.NewContext(r)
	_, err := GetUserByEmail(c, user.Email)

	if err != nil {
		_, err = user.Create(c, r)
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
			user := models.User{}
			user.GoogleId = userDetails["id"]
			user.Email = userDetails["email"]
			user.FirstName = userDetails["given_name"]
			user.LastName = userDetails["family_name"]
			user.EmailConfirmed = true
			_, err = user.Create(c, r)
		} else {
			context.Set(r, "user", user)
			Update(c, r, &user)
		}
	} else {
		user := context.Get(r, "user").(models.User)
		Update(c, r, &user)
	}
}

/*
* Update methods
 */

func Update(c appengine.Context, r *http.Request, u *models.User) (*models.User, error) {
	if len(u.Employers) == 0 {
		CreateAgencyFromUser(c, r, u)
	}
	return u, nil
}

func UpdateUser(c appengine.Context, r *http.Request, id string) (models.User, error) {
	// Get the details of the current user
	user, err := GetUser(c, r, id)
	if err != nil {
		return models.User{}, err
	}

	decoder := json.NewDecoder(r.Body)
	var updatedUser models.User
	err = decoder.Decode(&updatedUser)
	if err != nil {
		return models.User{}, err
	}

	UpdateIfNotBlank(&user.FirstName, updatedUser.FirstName)
	UpdateIfNotBlank(&user.LastName, updatedUser.LastName)

	if len(updatedUser.Employers) > 0 {
		user.Employers = updatedUser.Employers
	}

	user.Save(c)
	return user, nil
}
