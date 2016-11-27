package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/emails"
	"github.com/news-ai/web/google"
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

func filterEmailbyContactId(c context.Context, r *http.Request, contactId int64) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ContactId =", contactId)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil
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

func GetEmailById(c context.Context, r *http.Request, id int64) (models.Email, error) {
	email, err := getEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, err
	}
	return email, nil
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

	includedFiles := []models.File{}
	if len(email.Attachments) > 0 {
		for i := 0; i < len(email.Attachments); i++ {
			file, err := getFile(c, r, email.Attachments[i])
			if err == nil {
				includedFiles = append(includedFiles, file)
			} else {
				log.Errorf(c, "%v", err)
			}
		}
	}

	return email, includedFiles, nil
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
			_, err = emails[i].Create(c, r, currentUser)
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Email{}, nil, err
			}
			newEmails = append(newEmails, emails[i])
		}

		return newEmails, nil, err
	}

	// Create email
	_, err = email.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, err
	}
	return []models.Email{email}, nil, nil
}

func CreateEmailInternal(r *http.Request, to, firstName, lastName string) (models.Email, error) {
	c := appengine.NewContext(r)

	email := models.Email{}
	email.To = to
	email.FirstName = firstName
	email.LastName = lastName
	email.Created = time.Now()
	email.CreatedBy = int64(5749563331706880)

	_, err := email.Save(c)
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

func CancelEmail(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	// If it has not been delivered and has a sentat date then we can cancel it
	// and that sendAt date is in the future.
	if !email.Delievered && !email.SendAt.IsZero() && email.SendAt.After(time.Now()) {
		email.Cancel = true
		email.Save(c)
		return email, nil, nil
	}

	return email, nil, errors.New("Email has already been delivered")
}

func GetCurrentSchedueledEmails(c context.Context, r *http.Request) ([]models.Email, error) {
	emails := []models.Email{}

	timeNow := time.Now()
	currentTime := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), timeNow.Hour(), timeNow.Minute(), 0, 0, time.FixedZone("GMT", 0))

	// When the email is "Delievered == false" and has a "SendAt" date
	// And "Cancel == false". Also a filer if the user has sent it already or not
	query := datastore.NewQuery("Email").Filter("SendAt <=", currentTime).Filter("IsSent =", true).Filter("Delievered =", false).Filter("Cancel =", false)

	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil
}

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

	files := []models.File{}
	if len(email.Attachments) > 0 {
		for i := 0; i < len(email.Attachments); i++ {
			file, err := getFile(c, r, email.Attachments[i])
			if err == nil {
				files = append(files, file)
			} else {
				log.Errorf(c, "%v", err)
			}
		}
	}

	// Send through gmail
	if user.AccessToken != "" && user.Gmail {
		err = google.ValidateAccessToken(r, user)
		// Refresh access token if err is nil
		if err != nil {
			log.Errorf(c, "%v", err)
			user, err = google.RefreshAccessToken(r, user)
			if err != nil {
				log.Errorf(c, "%v", err)
				return email, nil, errors.New("Could not refresh user token")
			}
		}

		emailId := strconv.FormatInt(email.Id, 10)
		email.Body += "<img src=\"https://email2.newsai.co/?id=" + emailId + "\" alt=\"NewsAI\" />"

		email.Method = "gmail"
		val, err := email.MarkSent(c, "")
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		// Check to see if there is no sendat date or if date is in the past
		if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
			gmailId, gmailThreadId, err := emails.SendGmailEmail(r, user, email, files)
			if err != nil {
				log.Errorf(c, "%v", err)
				return email, nil, err
			}

			email.GmailId = gmailId
			email.GmailThreadId = gmailThreadId

			val, err = email.MarkDelivered(c)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}
		}

		return *val, nil, nil
	}

	email.Method = "sendgrid"
	val, err := email.MarkSent(c, "")
	if err != nil {
		log.Errorf(c, "%v", err)
		return *val, nil, err
	}

	// Check to see if there is no sendat date or if date is in the past
	if val.SendAt.IsZero() || val.SendAt.Before(time.Now()) {
		emailSent, emailId, err := emails.SendEmail(r, *val, user, files)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		val, err = email.MarkSent(c, emailId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		if emailSent {
			// Set attachments for deletion
			for i := 0; i < len(files); i++ {
				files[i].Imported = true
				files[i].Save(c)
			}

			return *val, nil, nil
		}
	}

	return *val, nil, nil
}

func MarkBounced(c context.Context, r *http.Request, e *models.Email, reason string) (*models.Email, models.NotificationChange, error) {
	SetUser(c, r, e.CreatedBy)
	notification, _ := LogNotificationForResource(c, r, "emails", e.Id, "BOUNCED", "")

	contacts, err := filterContactByEmail(c, e.To)
	if err != nil {
		log.Infof(c, "%v", err)
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].EmailBounced = true
		contacts[i].Save(c, r)
	}

	_, err = e.MarkBounced(c, reason)
	return e, notification, err
}

func MarkSpam(c context.Context, r *http.Request, e *models.Email) (*models.Email, models.NotificationChange, error) {
	SetUser(c, r, e.CreatedBy)
	notification, _ := LogNotificationForResource(c, r, "emails", e.Id, "SPAM", "")
	_, err := e.MarkSpam(c)
	return e, notification, err
}

func MarkClicked(c context.Context, r *http.Request, e *models.Email) (*models.Email, models.NotificationChange, error) {
	SetUser(c, r, e.CreatedBy)
	notification, _ := LogNotificationForResource(c, r, "emails", e.Id, "CLICKED", "")
	_, err := e.MarkClicked(c)
	return e, notification, err
}

func MarkDelivered(c context.Context, e *models.Email) (*models.Email, error) {
	return e.MarkDelivered(c)
}

func MarkOpened(c context.Context, r *http.Request, e *models.Email) (*models.Email, models.NotificationChange, error) {
	SetUser(c, r, e.CreatedBy)
	notification, _ := LogNotificationForResource(c, r, "emails", e.Id, "OPENED", "")
	_, err := e.MarkOpened(c)
	return e, notification, err
}
