package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Email struct {
	Id int64 `json:"id" datastore:"-"`

	Sender  string `json:"sender"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	IsSent bool `json:"issent"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

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

// Function to save a new email into App Engine
func (e *Email) save(c appengine.Context) (*Email, error) {
	// Update the Updated time
	e.Updated = time.Now()
	e.IsSent = false

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

func CreateEmailInternal(r *http.Request, to, firstName, lastName string) (Email, error) {
	c := appengine.NewContext(r)

	email := Email{}
	email.To = to
	email.FirstName = firstName
	email.LastName = lastName

	_, err := email.save(c)
	return email, err
}

/*
* Update methods
 */
