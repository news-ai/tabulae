package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/emails"
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

func getEmail(c context.Context, r *http.Request, id int64) (models.Email, error) {
	if id == 0 {
		return models.Email{}, errors.New("datastore: no such entity")
	}
	// Get the email by id
	var email models.Email
	emailId := datastore.NewKey(c, "Email", "", id, nil)
	err := nds.Get(c, emailId, &email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, err
	}

	if !email.Created.IsZero() {
		email.Format(emailId, "emails")

		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Email{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(email.CreatedBy, user.Id) {
			return models.Email{}, errors.New("Forbidden")
		}

		return email, nil
	}
	return models.Email{}, errors.New("No email by this id")
}

/*
* Filter methods
 */

func filterEmail(c context.Context, queryType, query string) (models.Email, error) {
	// Get a publication by the URL
	ks, err := datastore.NewQuery("Email").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, err
	}

	if len(ks) == 0 {
		return models.Email{}, errors.New("No email by the field " + queryType)
	}

	var emails []models.Email
	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, err
	}

	if len(emails) > 0 {
		emails[0].Format(ks[0], "emails")
		return emails[0], nil
	}
	return models.Email{}, errors.New("No email by this " + queryType)
}

func filterEmailbyListId(c context.Context, r *http.Request, listId int64) ([]models.Email, int, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, 0, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ListId =", listId)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, len(emails), nil
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil, len(emails), nil
}

func GetEmail(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	email, err := getEmail(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}
	return email, nil, nil
}

/*
* Create methods
 */

func CreateEmail(c context.Context, r *http.Request) ([]models.Email, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	var email models.Email
	err = decoder.Decode(buf, &email)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var emails []models.Email

		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &emails)

		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, err
		}

		newEmails := []models.Email{}
		for i := 0; i < len(emails); i++ {

			templateId := emails[i].TemplateId
			if templateId != 0 {
				singleTemplate, err := getTemplate(c, templateId)
				if err == nil {
					emails[i].Body = singleTemplate.Body
					emails[i].Subject = singleTemplate.Subject
				}
			}

			_, err = emails[i].Create(c, r, currentUser)
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Email{}, nil, err
			}
			// Logging the action happening
			LogNotificationForResource(c, r, "Email", emails[i].Id, "CREATE", "")
			newEmails = append(newEmails, emails[i])
		}

		return newEmails, nil, err
	}

	// Create email
	templateId := email.TemplateId
	if templateId != 0 {
		singleTemplate, err := getTemplate(c, templateId)
		if err == nil {
			email.Body = singleTemplate.Body
			email.Subject = singleTemplate.Subject
		}
	}
	_, err = email.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, err
	}

	// Logging the action happening
	LogNotificationForResource(c, r, "Email", email.Id, "CREATE", "")
	return []models.Email{email}, nil, nil
}

func CreateEmailInternal(r *http.Request, to, firstName, lastName string) (models.Email, error) {
	c := appengine.NewContext(r)

	email := models.Email{}
	email.To = to
	email.FirstName = firstName
	email.LastName = lastName

	_, err := email.Save(c)

	LogNotificationForResource(c, r, "Email", email.Id, "CREATE", "INTERNAL")
	return email, err
}

/*
* Filter methods
 */

func FilterEmailBySendGridID(c context.Context, sendGridId string) (models.Email, error) {
	// Get the id of the current email
	email, err := filterEmail(c, "SendGridId", sendGridId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, err
	}
	return email, nil
}

/*
* Update methods
 */

func UpdateEmail(c context.Context, r *http.Request, currentUser models.User, email *models.Email, updatedEmail models.Email) (models.Email, interface{}, error) {
	if email.CreatedBy != currentUser.Id {
		return *email, nil, errors.New("You don't have permissions to edit this object")
	}

	utilities.UpdateIfNotBlank(&email.Subject, updatedEmail.Subject)
	utilities.UpdateIfNotBlank(&email.Body, updatedEmail.Body)
	utilities.UpdateIfNotBlank(&email.To, updatedEmail.To)

	if updatedEmail.ListId != 0 {
		email.ListId = updatedEmail.ListId
	}

	if updatedEmail.TemplateId != 0 {
		email.TemplateId = updatedEmail.TemplateId
	}

	email.Save(c)

	// Logging the action happening
	LogNotificationForResource(c, r, "Email", email.Id, "UPDATE", "")

	return *email, nil, nil
}

func UpdateSingleEmail(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	// Get the details of the current email
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, errors.New("Could not get user")
	}

	if !permissions.AccessToObject(email.CreatedBy, user.Id) {
		return models.Email{}, nil, errors.New("Forbidden")
	}

	decoder := ffjson.NewDecoder()
	var updatedEmail models.Email
	buf, _ := ioutil.ReadAll(r.Body)
	err = decoder.Decode(buf, &updatedEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	return UpdateEmail(c, r, user, &email, updatedEmail)
}

func UpdateBatchEmail(c context.Context, r *http.Request) ([]models.Email, interface{}, error) {
	decoder := ffjson.NewDecoder()
	var updatedEmails []models.Email
	buf, _ := ioutil.ReadAll(r.Body)
	err := decoder.Decode(buf, &updatedEmails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, err
	}

	// Get logged in user
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, errors.New("Could not get user")
	}

	currentEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		email, err := getEmail(c, r, updatedEmails[i].Id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, err
		}

		if !permissions.AccessToObject(email.CreatedBy, user.Id) {
			return []models.Email{}, nil, errors.New("Forbidden")
		}

		currentEmails = append(currentEmails, email)
	}

	newEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		updatedEmail, _, err := UpdateEmail(c, r, user, &currentEmails[i], updatedEmails[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, err
		}
		newEmails = append(newEmails, updatedEmail)
	}

	return newEmails, nil, nil
}

/*
* Action methods
 */

func SendEmail(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return email, nil, err
	}

	if !user.EmailConfirmed {
		return email, nil, errors.New("Users email is not confirmed - the user cannot send emails.")
	}

	// Check if email is already sent
	if email.IsSent {
		return email, nil, errors.New("Email has already been sent.")
	}

	// Validate if HTML is valid
	validHTML := utilities.ValidateHTML(email.Body)
	if !validHTML {
		return email, nil, errors.New("Invalid HTML")
	}

	emailSent, emailId, err := emails.SendEmail(r, email, user)
	if err != nil {
		log.Errorf(c, "%v", err)
		return email, nil, err
	}

	if emailSent {
		val, err := email.MarkSent(c, emailId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		// Logging the action happening
		LogNotificationForResource(c, r, "Email", email.Id, "SENT", "")

		return *val, nil, nil
	}
	return email, nil, errors.New("Email could not be sent")
}
