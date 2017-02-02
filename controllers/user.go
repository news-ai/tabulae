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

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/sync"

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

// Gets every single user
func getUsersUnauthorized(c context.Context, r *http.Request) ([]models.User, error) {
	query := datastore.NewQuery("User")
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

func filterUserConfirmed(c context.Context, queryType, query string) (models.User, error) {
	// Get the current signed in user details by Id
	ks, err := datastore.NewQuery("User").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if len(ks) == 0 {
		return models.User{}, errors.New("No user by the field " + queryType)
	}

	// This shouldn't happen, but if the user double registers. Handle this case I guess
	if len(ks) > 1 {
		whichUserConfirmed := models.User{}
		for i := 0; i < len(ks); i++ {
			user := models.User{}
			userId := ks[i]

			err = nds.Get(c, userId, &user)
			if err != nil {
				log.Errorf(c, "%v", err)
				return models.User{}, err
			}

			if !user.Created.IsZero() {
				user.Format(userId, "users")

				if user.EmailConfirmed {
					whichUserConfirmed = user
				}
			}
		}

		// If none of them have confirmed their emails
		if whichUserConfirmed.Email == "" {
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
		}

		return whichUserConfirmed, nil
	} else {
		// The normal case where there's only one email
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

func GetUsersUnauthorized(c context.Context, r *http.Request) ([]models.User, error) {
	// Get the current user
	users, err := getUsersUnauthorized(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	return users, nil
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

func GetUserByEmailForValidation(c context.Context, email string) (models.User, error) {
	// Get the current user
	user, err := filterUserConfirmed(c, "Email", email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
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

		// Have a backup ConfirmationCode
		userBackup, errBackup := filterUser(c, "ConfirmationCodeBackup", confirmationCode)
		if errBackup != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		return userBackup, nil
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

func GetUserByIdUnauthorized(c context.Context, r *http.Request, userId int64) (models.User, error) {
	// Method dangerous since it can log into as any user. Be careful.
	user, err := getUserUnauthorized(c, r, userId)
	if err != nil {
		return models.User{}, nil
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

		sync.ResourceSync(r, user.Id, "User", "create")

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

	if user.RefreshToken != "" {
		existingUser.RefreshToken = user.RefreshToken
	}
	existingUser.TokenType = user.TokenType
	existingUser.GoogleExpiresIn = user.GoogleExpiresIn
	existingUser.Gmail = user.Gmail
	existingUser.GoogleId = user.GoogleId
	existingUser.AccessToken = user.AccessToken
	existingUser.GoogleCode = user.GoogleCode
	existingUser.Save(c)

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

func AddEmailToUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	// Only available when using SendGrid
	if user.Gmail || user.ExternalEmail {
		return user, nil, errors.New("Feature only works when using Sendgrid")
	}

	// Generate User Emails Code to send to confirmation email

	// Send Confirmation Email to this email address

	return user, nil, nil
}

func FeedbackFromUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var userFeedback models.UserFeedback
	err = decoder.Decode(buf, &userFeedback)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	// Get user's billing profile and add reasons there
	userBilling, err := GetUserBilling(c, r, currentUser)
	userBilling.ReasonNotPurchase = userFeedback.ReasonNotPurchase
	userBilling.FeedbackAfterTrial = userFeedback.FeedbackAfterTrial
	userBilling.Save(c)

	// Set the trial feedback to true - since they gave us feedback now
	user.TrialFeedback = true
	user.Save(c)

	sync.ResourceSync(r, user.Id, "User", "create")
	return user, nil, nil
}

/*
* Update methods
 */

func SaveUser(c context.Context, r *http.Request, u *models.User) (*models.User, error) {
	u.Save(c)
	sync.ResourceSync(r, u.Id, "User", "create")
	return u, nil
}

func Update(c context.Context, r *http.Request, u *models.User) (*models.User, error) {
	if len(u.Employers) == 0 {
		CreateAgencyFromUser(c, r, u)
	}

	billing, err := GetUserBilling(c, r, *u)
	if err != nil {
		return u, err
	}

	if billing.Expires.Before(time.Now()) {
		u.IsActive = false
		u.Save(c)

		billing.IsOnTrial = false
		billing.Save(c)
	}

	return u, nil
}

func UpdateUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
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
	utilities.UpdateIfNotBlank(&user.EmailSignature, updatedUser.EmailSignature)

	// If new user wants to get daily emails
	if updatedUser.GetDailyEmails == true {
		user.GetDailyEmails = true
	}

	// If this person doesn't want to get daily emails anymore
	if user.GetDailyEmails == true && updatedUser.GetDailyEmails == false {
		user.GetDailyEmails = false
	}

	if user.SMTPValid {
		// If new user wants to get daily emails
		if updatedUser.ExternalEmail == true {
			user.ExternalEmail = true
		}

		// If this person doesn't want to get daily emails anymore
		if user.ExternalEmail == true && updatedUser.ExternalEmail == false {
			user.ExternalEmail = false
		}
	}

	if len(updatedUser.Employers) > 0 {
		user.Employers = updatedUser.Employers
	}

	user.Save(c)
	sync.ResourceSync(r, user.Id, "User", "create")
	return user, nil, nil
}

/*
* Action methods
 */

func ValidateUserPassword(r *http.Request, email string, password string) (models.User, bool, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByEmailForValidation(c, email)
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

func SetUser(c context.Context, r *http.Request, userId int64) (models.User, error) {
	// Method dangerous since it can log into as any user. Be careful.
	user, err := getUserUnauthorized(c, r, userId)
	if err != nil {
		log.Errorf(c, "%v", err)
	}
	gcontext.Set(r, "user", user)
	return user, nil
}
