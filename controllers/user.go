package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/permissions"
	"github.com/news-ai/tabulae/utils"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getUser(c context.Context, r *http.Request, id int64) (models.User, error) {
	// Get the current signed in user details by Id
	var user models.User
	userId := datastore.NewKey(c, "User", "", id, nil)
	err := nds.Get(c, userId, &user)

	if err != nil {
		return models.User{}, err
	}

	if user.Email != "" {
		user.Id = userId.IntID()

		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		if !permissions.AccessToObject(user.Id, currentUser.Id) && !user.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		return user, nil
	}

	return models.User{}, errors.New("No user by this id")
}

// Gets every single user
func getUsers(c context.Context) ([]models.User, error) {
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

func filterUser(c context.Context, queryType, query string) (models.User, error) {
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

func GetUsers(c context.Context) ([]models.User, error) {
	// Get the current user
	users, err := getUsers(c)
	if err != nil {
		return []models.User{}, err
	}
	return users, nil
}

func GetUser(c context.Context, r *http.Request, id string) (models.User, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.User{}, err
		}
		return user, err
	default:
		userId, err := utils.StringIdToInt(id)
		if err != nil {
			return models.User{}, err
		}
		user, err := getUser(c, r, userId)
		if err != nil {
			return models.User{}, errors.New("No user by this id")
		}
		return user, nil
	}
}

func GetUserByEmail(c context.Context, email string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "Email", email)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByApiKey(c context.Context, apiKey string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ApiKey", apiKey)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByConfirmationCode(c context.Context, confirmationCode string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ConfirmationCode", confirmationCode)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetCurrentUser(c context.Context, r *http.Request) (models.User, error) {
	// Get the current user
	_, ok := gcontext.GetOk(r, "user")
	if !ok {
		return models.User{}, errors.New("No user logged in")
	}
	user := gcontext.Get(r, "user").(models.User)
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

func NewOrUpdateUser(c context.Context, r *http.Request, email string, userDetails map[string]string) {
	_, ok := gcontext.GetOk(r, "user")
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

			CreateNotificationForUser(c, r)
		} else {
			gcontext.Set(r, "user", user)
			Update(c, r, &user)
		}
	} else {
		user := gcontext.Get(r, "user").(models.User)
		Update(c, r, &user)
	}
}

func GetUserFromApiKey(r *http.Request, ApiKey string) (models.User, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByApiKey(c, ApiKey)

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

/*
* Update methods
 */

func Update(c context.Context, r *http.Request, u *models.User) (*models.User, error) {
	if len(u.Employers) == 0 {
		CreateAgencyFromUser(c, r, u)
	}
	return u, nil
}

func UpdateUser(c context.Context, r *http.Request, id string) (models.User, error) {
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

	utils.UpdateIfNotBlank(&user.FirstName, updatedUser.FirstName)
	utils.UpdateIfNotBlank(&user.LastName, updatedUser.LastName)

	if len(updatedUser.Employers) > 0 {
		user.Employers = updatedUser.Employers
	}

	user.Save(c)
	return user, nil
}
