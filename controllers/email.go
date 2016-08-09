package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/news-ai/tabulae/emails"
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

func getEmail(c context.Context, r *http.Request, id int64) (models.Email, error) {
	// Get the email by id
	emails := []models.Email{}
	emailId := datastore.NewKey(c, "Email", "", id, nil)
	ks, err := datastore.NewQuery("Email").Filter("__key__ =", emailId).GetAll(c, &emails)
	if err != nil {
		return models.Email{}, err
	}
	if len(emails) > 0 {
		emails[0].Id = ks[0].IntID()

		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.Email{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(emails[0].CreatedBy, user.Id) {
			return models.Email{}, errors.New("Forbidden")
		}

		return emails[0], nil
	}
	return models.Email{}, errors.New("No email by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetEmails(c context.Context, r *http.Request) ([]models.Email, error) {
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

func GetEmail(c context.Context, r *http.Request, id string) (models.Email, error) {
	// Get the details of the current user
	currentId, err := utils.StringIdToInt(id)
	if err != nil {
		return models.Email{}, err
	}

	email, err := getEmail(c, r, currentId)
	if err != nil {
		return models.Email{}, err
	}
	return email, nil
}

/*
* Create methods
 */

func CreateEmail(c context.Context, r *http.Request) ([]models.Email, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))

	decoder := json.NewDecoder(rdr1)
	var email models.Email
	err := decoder.Decode(&email)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Email{}, err
	}

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
			_, err = emails[i].Create(c, r, currentUser)
			if err != nil {
				return []models.Email{}, err
			}
			newEmails = append(newEmails, emails[i])
		}

		return newEmails, err
	}

	// Create email
	_, err = email.Create(c, r, currentUser)
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

	_, err := email.Save(c)
	return email, err
}

/*
* Update methods
 */

func UpdateEmail(c context.Context, email *models.Email, updatedEmail models.Email) (models.Email, error) {
	if email.CreatedBy != updatedEmail.CreatedBy {
		return *email, errors.New("You don't have permissions to edit this object")
	}

	utils.UpdateIfNotBlank(&email.Subject, updatedEmail.Subject)
	utils.UpdateIfNotBlank(&email.Body, updatedEmail.Body)
	utils.UpdateIfNotBlank(&email.To, updatedEmail.To)

	if updatedEmail.ListId != 0 {
		email.ListId = updatedEmail.ListId
	}

	email.Save(c)
	return *email, nil
}

func UpdateSingleEmail(c context.Context, r *http.Request, id string) (models.Email, error) {
	// Get the details of the current email
	email, err := GetEmail(c, r, id)
	if err != nil {
		return models.Email{}, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return models.Email{}, errors.New("Could not get user")
	}

	if !permissions.AccessToObject(email.CreatedBy, user.Id) {
		return models.Email{}, errors.New("Forbidden")
	}

	decoder := json.NewDecoder(r.Body)
	var updatedEmail models.Email
	err = decoder.Decode(&updatedEmail)
	if err != nil {
		return models.Email{}, err
	}

	return UpdateEmail(c, &email, updatedEmail)
}

func UpdateBatchEmail(c context.Context, r *http.Request) ([]models.Email, error) {
	decoder := json.NewDecoder(r.Body)
	var updatedEmails []models.Email
	err := decoder.Decode(&updatedEmails)
	if err != nil {
		return []models.Email{}, err
	}

	// Get logged in user
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Email{}, errors.New("Could not get user")
	}

	currentEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		email, err := getEmail(c, r, updatedEmails[i].Id)
		if err != nil {
			return []models.Email{}, err
		}

		if !permissions.AccessToObject(email.CreatedBy, user.Id) {
			return []models.Email{}, errors.New("Forbidden")
		}

		currentEmails = append(currentEmails, email)
	}

	newEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		updatedEmail, err := UpdateEmail(c, &currentEmails[i], updatedEmails[i])
		if err != nil {
			return []models.Email{}, err
		}
		newEmails = append(newEmails, updatedEmail)
	}

	return newEmails, nil
}

/*
* Action methods
 */

func SendEmail(c context.Context, r *http.Request, id string) (models.Email, error) {
	email, err := GetEmail(c, r, id)
	if err != nil {
		return models.Email{}, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return email, err
	}

	// Check if email is already sent
	if email.IsSent {
		return email, errors.New("Email has already been sent.")
	}

	// Validate if HTML is valid
	validHTML := utils.ValidateHTML(email.Body)
	if !validHTML {
		return email, errors.New("Invalid HTML")
	}

	emailSent, emailId := emails.SendEmail(r, email, user)
	if emailSent {
		val, err := email.MarkSent(c, emailId)
		if err != nil {
			return *val, err
		}
		return *val, nil
	}
	return email, errors.New("Email could not be sent")
}
