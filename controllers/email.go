package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/emails"
	"github.com/news-ai/web/google"
	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
)

type cancelEmailsBulk struct {
	Emails []int64 `json:"emails"`
}

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

		if !permissions.AccessToObject(email.CreatedBy, user.Id) && !user.IsAdmin {
			return models.Email{}, errors.New("Forbidden")
		}

		return email, nil
	}
	return models.Email{}, errors.New("No email by this id")
}

func getEmailUnauthorized(c context.Context, r *http.Request, id int64) (models.Email, error) {
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

func filterOrderedEmailbyContactId(c context.Context, r *http.Request, contactId int64) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ContactId =", contactId).Filter("IsSent =", true).Order("-Created")
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

func filterEmailbyContactId(c context.Context, r *http.Request, contactId int64) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ContactId =", contactId).Filter("IsSent =", true)
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

func GetSentEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("IsSent =", true).Filter("Cancel =", false).Filter("Delievered =", true)
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

func GetEmailStats(c context.Context, r *http.Request) (interface{}, interface{}, int, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	timeseriesData, count, err := search.SearchEmailTimeseriesByUserId(c, r, user)
	return timeseriesData, nil, count, err
}

func GetScheduledEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
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

func GetTeamEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	return []models.Email{}, nil, 0, nil
}

func GetEmailById(c context.Context, r *http.Request, id int64) (models.Email, error) {
	email, err := getEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, err
	}
	return email, nil
}

func GetEmailUnauthorized(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	email, err := getEmailUnauthorized(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	return email, nil, nil
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

func CreateEmailTransition(c context.Context, r *http.Request) ([]models.Email, interface{}, error) {
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

		var keys []*datastore.Key
		emailIds := []int64{}

		for i := 0; i < len(emails); i++ {
			// Test if the email we are sending with is in the user's SendGridFrom or is their Email
			if emails[i].FromEmail != "" {
				userEmailValid := false
				if currentUser.Email == emails[i].FromEmail {
					userEmailValid = true
				}

				for x := 0; x < len(currentUser.Emails); x++ {
					if currentUser.Emails[x] == emails[i].FromEmail {
						userEmailValid = true
					}
				}

				// If this is if the email added is not valid in SendGridFrom
				if !userEmailValid {
					return []models.Email{}, nil, errors.New("The email requested is not confirmed by the user yet")
				}
			}

			emails[i].Id = 0
			emails[i].CreatedBy = currentUser.Id
			emails[i].Created = time.Now()
			emails[i].Updated = time.Now()
			emails[i].TeamId = currentUser.TeamId
			emails[i].IsSent = false

			keys = append(keys, emails[i].Key(c))
		}

		if len(keys) < 300 {
			ks := []*datastore.Key{}
			err = nds.RunInTransaction(c, func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
				ks, err = nds.PutMulti(contextWithTimeout, keys, emails)
				if err != nil {
					log.Errorf(c, "%v", err)
					return err
				}
				return nil
			}, nil)

			for i := 0; i < len(ks); i++ {
				emails[i].Format(ks[i], "emails")
				emailIds = append(emailIds, emails[i].Id)
			}

			sync.EmailResourceBulkSync(r, emailIds)
			return emails, nil, err
		} else {
			firstHalfKeys := []*datastore.Key{}
			secondHalfKeys := []*datastore.Key{}
			thirdHalfKeys := []*datastore.Key{}
			fourHalfKeys := []*datastore.Key{}

			startOne := 0
			endOne := 100

			startTwo := 100
			endTwo := 200

			startThree := 200
			endThree := 300

			startFour := 300
			endFour := len(keys)

			err1 := nds.RunInTransaction(c, func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
				firstHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startOne:endOne], emails[startOne:endOne])
				if err != nil {
					log.Errorf(c, "%v", err)
					return err
				}
				return nil
			}, nil)

			err2 := nds.RunInTransaction(c, func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
				secondHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startTwo:endTwo], emails[startTwo:endTwo])
				if err != nil {
					log.Errorf(c, "%v", err)
					return err
				}
				return nil
			}, nil)

			err3 := nds.RunInTransaction(c, func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
				thirdHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startThree:endThree], emails[startThree:endThree])
				if err != nil {
					log.Errorf(c, "%v", err)
					return err
				}
				return nil
			}, nil)

			err4 := nds.RunInTransaction(c, func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
				fourHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startFour:endFour], emails[startFour:endFour])
				if err != nil {
					log.Errorf(c, "%v", err)
					return err
				}
				return nil
			}, nil)

			firstHalfKeys = append(firstHalfKeys, secondHalfKeys...)
			firstHalfKeys = append(firstHalfKeys, thirdHalfKeys...)
			firstHalfKeys = append(firstHalfKeys, fourHalfKeys...)

			for i := 0; i < len(firstHalfKeys); i++ {
				emails[i].Format(firstHalfKeys[i], "emails")
				emailIds = append(emailIds, emails[i].Id)
			}

			if err1 != nil {
				err = err1
			}

			if err2 != nil {
				err = err2
			}

			if err3 != nil {
				err = err3
			}

			if err4 != nil {
				err = err4
			}

			sync.EmailResourceBulkSync(r, emailIds)
			return emails, nil, err
		}
	}

	// Test if the email we are sending with is in the user's SendGridFrom or is their Email
	if email.FromEmail != "" {
		userEmailValid := false
		if currentUser.Email == email.FromEmail {
			userEmailValid = true
		}

		for i := 0; i < len(currentUser.Emails); i++ {
			if currentUser.Emails[i] == email.FromEmail {
				userEmailValid = true
			}
		}

		// If this is if the email added is not valid in SendGridFrom
		if !userEmailValid {
			return []models.Email{}, nil, errors.New("The email requested is not confirmed by you yet")
		}
	}

	email.TeamId = currentUser.TeamId

	// Create email
	_, err = email.Create(c, r, currentUser)
	sync.ResourceSync(r, email.Id, "Email", "create")
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

	email.CC = updatedEmail.CC
	email.BCC = updatedEmail.BCC

	if updatedEmail.ListId != 0 {
		email.ListId = updatedEmail.ListId
	}

	if updatedEmail.TemplateId != 0 {
		email.TemplateId = updatedEmail.TemplateId
	}

	email.Save(c)
	sync.ResourceSync(r, email.Id, "Email", "create")
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

func CancelAllScheduled(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	emails := []models.Email{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
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

	emailIds := []int64{} // Validated email ids
	for i := 0; i < len(emails); i++ {
		// If it has not been delivered and has a sentat date then we can cancel it
		// and that sendAt date is in the future.
		if !emails[i].Delievered && !emails[i].SendAt.IsZero() && emails[i].SendAt.After(time.Now()) {
			emails[i].Cancel = true
			emails[i].Save(c)
			emailIds = append(emailIds, emails[i].Id)
		}
	}

	sync.EmailResourceBulkSync(r, emailIds)
	return emails, nil, len(emails), nil
}

func BulkCancelEmail(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var cancelEmails cancelEmailsBulk
	err := decoder.Decode(buf, &cancelEmails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	emails := []models.Email{}
	emailIds := []int64{} // Validated email ids
	for i := 0; i < len(cancelEmails.Emails); i++ {
		email, err := getEmail(c, r, cancelEmails.Emails[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			continue
		}

		// If it has not been delivered and has a sentat date then we can cancel it
		// and that sendAt date is in the future.
		if !email.Delievered && !email.SendAt.IsZero() && email.SendAt.After(time.Now()) {
			email.Cancel = true
			email.Save(c)
			emails = append(emails, email)
			emailIds = append(emailIds, email.Id)
		}
	}

	sync.EmailResourceBulkSync(r, emailIds)
	return emails, nil, len(emails), nil
}

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
		sync.ResourceSync(r, email.Id, "Email", "create")
		return email, nil, nil
	}

	return email, nil, errors.New("Email has already been delivered")
}

func ArchiveEmail(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	email.Archived = true
	email.Save(c)
	sync.ResourceSync(r, email.Id, "Email", "create")
	return email, nil, nil
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

	emailsToSend := []models.Email{}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")

		if !emails[i].SendAt.IsZero() {
			emailsToSend = append(emailsToSend, emails[i])
		}
	}

	return emailsToSend, nil
}

func BulkSendEmail(c context.Context, r *http.Request) ([]models.Email, interface{}, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var bulkEmailIds models.BulkSendEmailIds
	err := decoder.Decode(buf, &bulkEmailIds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	emails := []models.Email{}
	emailIds := []int64{}

	for i := 0; i < len(bulkEmailIds.EmailIds); i++ {
		singleEmail, _, err := SendEmail(c, r, strconv.FormatInt(bulkEmailIds.EmailIds[i], 10), false)
		if err != nil {
			log.Errorf(c, "%v", err)
		}
		emails = append(emails, singleEmail)
		emailIds = append(emailIds, singleEmail.Id)
	}

	sync.EmailResourceBulkSync(r, emailIds)
	return emails, nil, len(emails), nil
}

func SendEmail(c context.Context, r *http.Request, id string, isNotBulk bool) (models.Email, interface{}, error) {
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

	emailId := strconv.FormatInt(email.Id, 10)
	email.Body = utilities.AppendHrefWithLink(c, email.Body, emailId, "https://email2.newsai.co/a")
	email.Body += "<img src=\"https://email2.newsai.co/?id=" + emailId + "\" alt=\"NewsAI\" />"
	if user.SMTPValid && user.ExternalEmail && user.EmailSetting != 0 {
		email.Method = "smtp"
		val, err := email.MarkSent(c, "")
		if err != nil {
			log.Errorf(c, "%v", err)
			return *val, nil, err
		}

		// Check to see if there is no sendat date or if date is in the past
		if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
			emailBody, err := emails.GenerateEmail(r, user, email, files)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}

			emailSetting, err := getEmailSetting(c, r, user.EmailSetting)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}

			SMTPPassword := string(user.SMTPPassword[:])

			contextWithTimeout, _ := context.WithTimeout(c, time.Second*30)
			client := urlfetch.Client(contextWithTimeout)
			getUrl := "https://tabulae-smtp.newsai.org/send"

			sendEmailRequest := models.SMTPEmailSettings{}
			sendEmailRequest.Servername = emailSetting.SMTPServer + ":" + strconv.Itoa(emailSetting.SMTPPortSSL)
			sendEmailRequest.EmailUser = user.SMTPUsername
			sendEmailRequest.EmailPassword = SMTPPassword
			sendEmailRequest.To = email.To
			sendEmailRequest.Subject = email.Subject
			sendEmailRequest.Body = emailBody

			SendEmailRequest, err := json.Marshal(sendEmailRequest)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}
			log.Infof(c, "%v", string(SendEmailRequest))
			sendEmailQuery := bytes.NewReader(SendEmailRequest)

			req, _ := http.NewRequest("POST", getUrl, sendEmailQuery)

			resp, err := client.Do(req)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}

			decoder := json.NewDecoder(resp.Body)
			var verifyResponse SMTPEmailResponse
			err = decoder.Decode(&verifyResponse)
			if err != nil {
				log.Errorf(c, "%v", err)
				return *val, nil, err
			}

			log.Infof(c, "%v", verifyResponse)

			if verifyResponse.Status {
				val, err = email.MarkDelivered(c)
				if err != nil {
					log.Errorf(c, "%v", err)
					return *val, nil, err
				}
				return *val, nil, nil
			}

			if isNotBulk {
				sync.ResourceSync(r, val.Id, "Email", "create")
			}
			return *val, nil, errors.New(verifyResponse.Error)
		}

		return *val, nil, nil
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

		if isNotBulk {
			sync.ResourceSync(r, val.Id, "Email", "create")
		}
		return *val, nil, nil
	}

	// if user.IsAdmin {
	// 	// Use SparkPost
	// 	log.Infof(c, "%v", "Using SparkPost")

	// 	// Mark email as sent again with "sparkpost" method
	// 	email.Method = "sparkpost"
	// 	val, err := email.MarkSent(c, "")
	// 	if err != nil {
	// 		log.Errorf(c, "%v", err)
	// 		return *val, nil, err
	// 	}

	// 	// Test if the email we are sending with is in the user's SendGridFrom or is their Email
	// 	if val.FromEmail != "" {
	// 		userEmailValid := false
	// 		if user.Email == val.FromEmail {
	// 			userEmailValid = true
	// 		}

	// 		for i := 0; i < len(user.Emails); i++ {
	// 			if user.Emails[i] == val.FromEmail {
	// 				userEmailValid = true
	// 			}
	// 		}

	// 		// If this is if the email added is not valid in SendGridFrom
	// 		if !userEmailValid {
	// 			return *val, nil, errors.New("The email requested is not confirmed by the user yet")
	// 		}
	// 	}

	// 	// Check to see if there is no sendat date or if date is in the past
	// 	if val.SendAt.IsZero() || val.SendAt.Before(time.Now()) {
	// 		emailSent, emailId, err := emails.SendSparkPostEmail(r, *val, user, files)
	// 		if err != nil {
	// 			log.Errorf(c, "%v", err)
	// 			return *val, nil, err
	// 		}

	// 		val.SparkPostId = emailId
	// 		val, err = email.MarkSent(c, "")
	// 		if err != nil {
	// 			log.Errorf(c, "%v", err)
	// 			return *val, nil, err
	// 		}

	// 		val, err = email.MarkDelivered(c)
	// 		if err != nil {
	// 			log.Errorf(c, "%v", err)
	// 			return *val, nil, err
	// 		}

	// 		if emailSent {
	// 			// Set attachments for deletion
	// 			for i := 0; i < len(files); i++ {
	// 				files[i].Imported = true
	// 				files[i].Save(c)
	// 			}

	// 			if isNotBulk {
	// 				sync.ResourceSync(r, val.Id, "Email", "create")
	// 			}
	// 			return *val, nil, nil
	// 		}
	// 	}
	// }

	email.Method = "sendgrid"
	val, err := email.MarkSent(c, "")
	if err != nil {
		log.Errorf(c, "%v", err)
		return *val, nil, err
	}

	// Test if the email we are sending with is in the user's SendGridFrom or is their Email
	if val.FromEmail != "" {
		userEmailValid := false
		if user.Email == val.FromEmail {
			userEmailValid = true
		}

		for i := 0; i < len(user.Emails); i++ {
			if user.Emails[i] == val.FromEmail {
				userEmailValid = true
			}
		}

		// If this is if the email added is not valid in SendGridFrom
		if !userEmailValid {
			return *val, nil, errors.New("The email requested is not confirmed by the user yet")
		}
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

		val, err = email.MarkDelivered(c)
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

			if isNotBulk {
				sync.ResourceSync(r, val.Id, "Email", "create")
			}
			return *val, nil, nil
		}
	}

	if isNotBulk {
		sync.ResourceSync(r, val.Id, "Email", "create")
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

func MarkDelivered(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	_, err := e.MarkDelivered(c)
	return e, err
}

func MarkOpened(c context.Context, r *http.Request, e *models.Email) (*models.Email, models.NotificationChange, error) {
	SetUser(c, r, e.CreatedBy)
	notification, _ := LogNotificationForResource(c, r, "emails", e.Id, "OPENED", "")
	_, err := e.MarkOpened(c)
	return e, notification, err
}

func GetEmailLogs(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
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

	logs, _, err := search.SearchEmailLogByEmailId(c, r, user, email.Id)
	return logs, nil, err
}

func GetEmailSearch(c context.Context, r *http.Request) (interface{}, interface{}, int, error) {
	queryField := gcontext.Get(r, "q").(string)

	if queryField == "" {
		return nil, nil, 0, nil
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	if strings.Contains(queryField, "date:") {
		emailDate := strings.Split(queryField, "date:")
		log.Infof(c, "%v", emailDate)
		if len(emailDate) > 1 {
			emails, count, err := search.SearchEmailsByDate(c, r, user, emailDate[1])
			return emails, nil, count, err
		} else {
			return nil, nil, 0, errors.New("Please enter a valid date")
		}
	}

	emails, count, err := search.SearchEmailsByQuery(c, r, user, queryField)
	return emails, nil, count, err
}
