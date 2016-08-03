package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/news-ai/tabulae/models"
)

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (e *models.Email) key(c appengine.Context) *datastore.Key {
	if e.Id == 0 {
		e.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Email", nil)
	}
	return datastore.NewKey(c, "Email", "", e.Id, nil)
}

/*
* Get methods
 */

func getEmail(c appengine.Context, id int64) (models.Email, error) {
	// Get the email by id
	emails := []models.Email{}
	emailId := datastore.NewKey(c, "Email", "", id, nil)
	ks, err := datastore.NewQuery("Email").Filter("__key__ =", emailId).GetAll(c, &emails)
	if err != nil {
		return models.Email{}, err
	}
	if len(emails) > 0 {
		emails[0].Id = ks[0].IntID()
		return emails[0], nil
	}
	return models.Email{}, errors.New("No email by this id")
}

/*
* Create methods
 */

func (e *models.Email) create(c appengine.Context, r *http.Request) (*models.Email, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return e, err
	}

	e.IsSent = false
	e.CreatedBy = currentUser.Id
	e.Created = time.Now()

	_, err = e.save(c)
	return e, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (e *models.Email) save(c appengine.Context) (*models.Email, error) {
	// Update the Updated time
	e.Updated = time.Now()

	k, err := datastore.Put(c, e.key(c), e)
	if err != nil {
		return nil, err
	}
	e.Id = k.IntID()
	return e, nil
}

func (e *models.Email) MarkSent(c appengine.Context) (*models.Email, error) {
	e.IsSent = true
	_, err := e.save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetEmails(c appengine.Context, r *http.Request) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Email{}, err
	}

	ks, err := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).GetAll(c, &emails)
	if err != nil {
		return []models.Email{}, err
	}
	for i := 0; i < len(emails); i++ {
		emails[i].Id = ks[i].IntID()
	}

	return emails, nil
}

func GetEmail(c appengine.Context, id string) (models.Email, error) {
	// Get the details of the current user
	currentId, err := StringIdToInt(id)
	if err != nil {
		return models.Email{}, err
	}

	email, err := getEmail(c, currentId)
	if err != nil {
		return models.Email{}, err
	}
	return email, nil
}

/*
* Create methods
 */

func CreateEmail(c appengine.Context, r *http.Request) ([]models.Email, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))

	decoder := json.NewDecoder(rdr1)
	var email models.Email
	err := decoder.Decode(&email)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var emails []models.Email

		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
		arrayDecoder := json.NewDecoder(rdr2)
		err = arrayDecoder.Decode(&emails)

		if err != nil {
			return []models.Email{}, err
		}

		newEmails := []models.Email{}
		for i := 0; i < len(emails); i++ {
			_, err = emails[i].create(c, r)
			if err != nil {
				return []models.Email{}, err
			}
			newEmails = append(newEmails, emails[i])
		}

		return newEmails, err
	}

	// Create email
	_, err = email.create(c, r)
	if err != nil {
		return []models.Email{}, err
	}

	return []models.Email{email}, nil
}

func CreateEmailInternal(r *http.Request, to, firstName, lastName string) (models.Email, error) {
	c := appengine.NewContext(r)

	email := models.Email{}
	email.To = to
	email.FirstName = firstName
	email.LastName = lastName

	_, err := email.save(c)
	return email, err
}

/*
* Update methods
 */

func UpdateEmail(c appengine.Context, email *models.Email, updatedEmail models.Email) models.Email {
	UpdateIfNotBlank(&email.Subject, updatedEmail.Subject)
	UpdateIfNotBlank(&email.Body, updatedEmail.Body)
	UpdateIfNotBlank(&email.To, updatedEmail.To)

	if updatedEmail.ListId != 0 {
		email.ListId = updatedEmail.ListId
	}

	email.save(c)
	return *email
}

func UpdateSingleEmail(c appengine.Context, r *http.Request, id string) (models.Email, error) {
	// Get the details of the current email
	email, err := GetEmail(c, id)
	if err != nil {
		return models.Email{}, err
	}

	decoder := json.NewDecoder(r.Body)
	var updatedEmail models.Email
	err = decoder.Decode(&updatedEmail)
	if err != nil {
		return models.Email{}, err
	}

	return UpdateEmail(c, &email, updatedEmail), nil
}

func UpdateBatchEmail(c appengine.Context, r *http.Request) ([]models.Email, error) {
	decoder := json.NewDecoder(r.Body)
	var updatedEmails []models.Email
	err := decoder.Decode(&updatedEmails)
	if err != nil {
		return []models.Email{}, err
	}

	newEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		email, err := getEmail(c, updatedEmails[i].Id)
		if err != nil {
			return []models.Email{}, err
		}
		updatedEmail := UpdateEmail(c, &email, updatedEmails[i])
		newEmails = append(newEmails, updatedEmail)
	}

	return newEmails, nil
}
