package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type EmailUser struct {
	Id int64  `json:"id" datastore:"-"`
	To string `json:"to"`

	SeenAt []time.Time `json:"seenat"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Email struct {
	Id int64 `json:"id" datastore:"-"`

	Sender  string  `json:"sender"`
	To      []int64 `json:"to"`
	Subject string  `json:"subject"`
	Body    string  `json:"body"`

	// User details
	FirstName string `json:"firstname"`

	IsSent bool `json:"issent"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (eu *EmailUser) key(c appengine.Context) *datastore.Key {
	if eu.Id == 0 {
		eu.Created = time.Now()
		return datastore.NewIncompleteKey(c, "EmailUser", nil)
	}
	return datastore.NewKey(c, "EmailUser", "", eu.Id, nil)
}

// Generates a new key for the data to be stored on App Engine
func (e *Email) key(c appengine.Context) *datastore.Key {
	if e.Id == 0 {
		e.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Email", nil)
	}
	return datastore.NewKey(c, "Email", "", e.Id, nil)
}

/*
* Get methods
 */

func getEmailUser(c appengine.Context, id int64) (EmailUser, error) {
	// Get the Contact by id
	emailUsers := []EmailUser{}
	emailUserId := datastore.NewKey(c, "EmailUser", "", id, nil)
	ks, err := datastore.NewQuery("EmailUser").Filter("__key__ =", emailUserId).GetAll(c, &emailUsers)
	if err != nil {
		return EmailUser{}, err
	}
	if len(emailUsers) > 0 {
		emailUsers[0].Id = ks[0].IntID()
		return emailUsers[0], nil
	}
	return EmailUser{}, errors.New("No emailuser by this id")
}

func getEmail(c appengine.Context, id int64) (Email, error) {
	// Get the Contact by id
	emails := []Email{}
	emailId := datastore.NewKey(c, "Email", "", id, nil)
	ks, err := datastore.NewQuery("Email").Filter("__key__ =", emailId).GetAll(c, &emails)
	if err != nil {
		return Email{}, err
	}
	if len(emails) > 0 {
		emails[0].Id = ks[0].IntID()
		return emails[0], nil
	}
	return Email{}, errors.New("No email by this id")
}

/*
* Create methods
 */

func (eu *EmailUser) create(c appengine.Context, r *http.Request) (*EmailUser, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return eu, err
	}

	eu.CreatedBy = currentUser.Id
	eu.Created = time.Now()

	_, err = eu.save(c)
	return eu, err
}

func (e *Email) create(c appengine.Context, r *http.Request) (*Email, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return e, err
	}

	e.CreatedBy = currentUser.Id
	e.Created = time.Now()

	_, err = e.save(c)
	return e, err
}

/*
* Update methods
 */

// Function to save a new emailuser into App Engine
func (eu *EmailUser) save(c appengine.Context) (*EmailUser, error) {
	// Update the Updated time
	eu.Updated = time.Now()

	k, err := datastore.Put(c, eu.key(c), eu)
	if err != nil {
		return nil, err
	}
	eu.Id = k.IntID()
	return eu, nil
}

// Function to save a new email into App Engine
func (e *Email) save(c appengine.Context) (*Email, error) {
	// Update the Updated time
	e.Updated = time.Now()

	k, err := datastore.Put(c, e.key(c), e)
	if err != nil {
		return nil, err
	}
	e.Id = k.IntID()
	return e, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single email users
func GetEmailUsers(c appengine.Context, r *http.Request) ([]EmailUser, error) {
	emailUsers := []EmailUser{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []EmailUser{}, err
	}

	ks, err := datastore.NewQuery("EmailUser").Filter("CreatedBy =", user.Id).GetAll(c, &emailUsers)
	if err != nil {
		return []EmailUser{}, err
	}
	for i := 0; i < len(emailUsers); i++ {
		emailUsers[i].Id = ks[i].IntID()
	}

	return emailUsers, nil
}

func GetEmails(c appengine.Context, r *http.Request) ([]Email, error) {
	emails := []Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []Email{}, err
	}

	ks, err := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).GetAll(c, &emails)
	if err != nil {
		return []Email{}, err
	}
	for i := 0; i < len(emails); i++ {
		emails[i].Id = ks[i].IntID()
	}

	return emails, nil
}

func GetEmailUser(c appengine.Context, id string) (EmailUser, error) {
	// Get the details of the current user
	currentId, err := StringIdToInt(id)
	if err != nil {
		return EmailUser{}, err
	}

	emailUser, err := getEmailUser(c, currentId)
	if err != nil {
		return EmailUser{}, err
	}
	return emailUser, nil
}

func GetEmail(c appengine.Context, id string) (Email, error) {
	// Get the details of the current user
	currentId, err := StringIdToInt(id)
	if err != nil {
		return Email{}, err
	}

	email, err := getEmail(c, currentId)
	if err != nil {
		return Email{}, err
	}
	return email, nil
}

/*
* Create methods
 */

func CreateEmailUser(c appengine.Context, w http.ResponseWriter, r *http.Request) (EmailUser, error) {
	decoder := json.NewDecoder(r.Body)
	var emailUser EmailUser
	err := decoder.Decode(&emailUser)
	if err != nil {
		return EmailUser{}, err
	}

	// Create email user
	_, err = emailUser.create(c, r)
	if err != nil {
		return EmailUser{}, err
	}

	return emailUser, nil
}

func CreateEmailUserInternal(r *http.Request, email string) (EmailUser, error) {
	c := appengine.NewContext(r)

	emailUser := EmailUser{}
	emailUser.To = email

	_, err := emailUser.save(c)
	return emailUser, err
}

func CreateEmail(c appengine.Context, w http.ResponseWriter, r *http.Request) (Email, error) {
	decoder := json.NewDecoder(r.Body)
	var email Email
	err := decoder.Decode(&email)
	if err != nil {
		return Email{}, err
	}

	// Create email
	_, err = email.create(c, r)
	if err != nil {
		return Email{}, err
	}

	return email, nil
}

func CreateEmailInternal(r *http.Request, to []int64) (Email, error) {
	c := appengine.NewContext(r)

	email := Email{}
	email.To = to

	_, err := email.save(c)
	return email, err
}

/*
* Update methods
 */
