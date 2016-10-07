package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/billing"
	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
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
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if user.Email != "" {
		user.Format(userId, "users")
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		return user, nil
	}
	return models.User{}, errors.New("No user by this id")
}

func getUserUnauthorized(c context.Context, r *http.Request, id int64) (models.User, error) {
	// Get the current signed in user details by Id
	var user models.User
	userId := datastore.NewKey(c, "User", "", id, nil)
	err := nds.Get(c, userId, &user)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if user.Email != "" {
		user.Format(userId, "users")
		return user, nil
	}
	return models.User{}, errors.New("No user by this id")
}

// Gets every single user
func getUsers(c context.Context, r *http.Request) ([]models.User, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	if !user.IsAdmin {
		return []models.User{}, errors.New("Forbidden")
	}

	query := datastore.NewQuery("User")
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	var users []models.User
	users = make([]models.User, len(ks))
	err = nds.GetMulti(c, ks, users)
	if err != nil {
		log.Infof(c, "%v", err)
		return users, err
	}

	for i := 0; i < len(users); i++ {
		users[i].Format(ks[i], "users")
	}
	return users, nil
}

/*
* Filter methods
 */

func filterUser(c context.Context, queryType, query string) (models.User, error) {
	// Get the current signed in user details by Id
	ks, err := datastore.NewQuery("User").Filter(queryType+" =", query).Limit(1).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if len(ks) == 0 {
		return models.User{}, errors.New("No user by the field " + queryType)
	}

	user := models.User{}
	userId := ks[0]

	err = nds.Get(c, userId, &user)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if !user.Created.IsZero() {
		user.Format(userId, "users")
		return user, nil
	}
	return models.User{}, errors.New("No user by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetUsers(c context.Context, r *http.Request) ([]models.User, interface{}, int, error) {
	// Get the current user
	users, err := getUsers(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, nil, 0, err
	}

	return users, nil, len(users), nil
}

func GetUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		return user, nil, err
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err := getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		return user, nil, nil
	}
}

func GetUserByEmail(c context.Context, email string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "Email", email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByApiKey(c context.Context, apiKey string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ApiKey", apiKey)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByConfirmationCode(c context.Context, confirmationCode string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ConfirmationCode", confirmationCode)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByResetCode(c context.Context, resetCode string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ResetPasswordCode", resetCode)
	if err != nil {
		log.Errorf(c, "%v", err)
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

func GetUserFromApiKey(r *http.Request, ApiKey string) (models.User, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByApiKey(c, ApiKey)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

/*
* Create methods
 */

func RegisterUser(r *http.Request, user models.User) (models.User, bool, error) {
	c := appengine.NewContext(r)
	existingUser, err := GetUserByEmail(c, user.Email)

	if err != nil {
		// Validation if the email is null
		if user.Email == "" {
			noEmailErr := errors.New("User does have an email")
			log.Errorf(c, "%v", noEmailErr)
			log.Errorf(c, "%v", user)
			return models.User{}, false, noEmailErr
		}

		// Add the user to datastore
		_, err = user.Create(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return user, false, err
		}

		// Set the user
		gcontext.Set(r, "user", user)
		Update(c, r, &user)

		// Create a sample media list for the user
		_, _, err = CreateSampleMediaList(c, r, user)
		if err != nil {
			log.Errorf(c, "%v", err)
		}
		return user, true, nil
	}
	return existingUser, false, errors.New("User with the email already exists")
}

func AddUserToContext(c context.Context, r *http.Request, email string) {
	_, ok := gcontext.GetOk(r, "user")
	if !ok {
		user, _ := GetUserByEmail(c, email)
		gcontext.Set(r, "user", user)
		Update(c, r, &user)
	} else {
		user := gcontext.Get(r, "user").(models.User)
		Update(c, r, &user)
	}
}

/*
* Update methods
 */

func Update(c context.Context, r *http.Request, u *models.User) (*models.User, error) {
	if len(u.Employers) == 0 {
		CreateAgencyFromUser(c, r, u)
	}
	if u.StripeId == "" {
		if time.Now().Month().String() == "October" {
			billing.CreateBetaCustomer(r, *u)
		} else {
			billing.CreateCustomer(r, *u)
		}
	}
	return u, nil
}

func UpdateUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	// Get the details of the current user
	user, _, err := GetUser(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedUser models.User
	err = decoder.Decode(buf, &updatedUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	utilities.UpdateIfNotBlank(&user.FirstName, updatedUser.FirstName)
	utilities.UpdateIfNotBlank(&user.LastName, updatedUser.LastName)

	if len(updatedUser.Employers) > 0 {
		user.Employers = updatedUser.Employers
	}

	user.Save(c)
	return user, nil, nil
}

/*
* Action methods
 */

func ValidateUserPassword(r *http.Request, email string, password string) (models.User, bool, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByEmail(c, email)
	if err == nil {
		err = utilities.ValidatePassword(user.Password, password)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, false, nil
		}
		return user, true, nil
	}
	return models.User{}, false, errors.New("User does not exist")
}

/*
* Action methods
 */

func SetUser(c context.Context, r *http.Request, userId int64) (models.User, error) {
	// Method dangerous since it can log into as any user. Be careful.
	user, err := getUserUnauthorized(c, r, userId)
	if err != nil {
		log.Errorf(c, "%v", err)
	}
	gcontext.Set(r, "user", user)
	return user, nil
}
