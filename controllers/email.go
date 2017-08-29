package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/api/controllers"
	apiModels "github.com/news-ai/api/models"

	"github.com/news-ai/tabulae/attach"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"

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

		user, err := controllers.GetCurrentUser(c, r)
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

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, 0, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ListId =", listId)
	query = controllers.ConstructQuery(query, r)
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

func filterOrderedEmailbyContactId(c context.Context, r *http.Request, contact models.Contact) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("To =", contact.Email).Filter("IsSent =", true).Filter("Cancel =", false).Filter("Archived =", false).Order("-Created")
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

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ContactId =", contactId).Filter("IsSent =", true)
	query = controllers.ConstructQuery(query, r)
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

func filterEmailbyContactEmail(c context.Context, r *http.Request, email string) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("To =", email).Filter("IsSent =", true)
	query = controllers.ConstructQuery(query, r)
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

func emailsToLists(c context.Context, r *http.Request, emails []models.Email) []models.MediaList {
	mediaListIds := []int64{}

	for i := 0; i < len(emails); i++ {
		if emails[i].ListId != 0 {
			mediaListIds = append(mediaListIds, emails[i].ListId)
		}
	}

	// Work on includes
	mediaLists := []models.MediaList{}
	mediaListExists := map[int64]bool{}
	mediaListExists = make(map[int64]bool)

	for i := 0; i < len(mediaListIds); i++ {
		if _, ok := mediaListExists[mediaListIds[i]]; !ok {
			if mediaListIds[i] != 0 {
				mediaList, _ := getMediaList(c, r, mediaListIds[i])
				mediaLists = append(mediaLists, mediaList)
				mediaListExists[mediaListIds[i]] = true
			}
		}
	}

	return mediaLists
}

func emailsToContacts(c context.Context, r *http.Request, emails []models.Email) []models.Contact {
	contactIds := []int64{}

	for i := 0; i < len(emails); i++ {
		if emails[i].ContactId != 0 {
			contactIds = append(contactIds, emails[i].ContactId)
		}
	}

	// Work on includes
	contacts := []models.Contact{}
	contactExists := map[int64]bool{}
	contactExists = make(map[int64]bool)

	for i := 0; i < len(contactIds); i++ {
		if _, ok := contactExists[contactIds[i]]; !ok {
			if contactIds[i] != 0 {
				contact, _ := getContact(c, r, contactIds[i])
				contacts = append(contacts, contact)
				contactExists[contactIds[i]] = true
			}
		}
	}

	return contacts
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(c, r, emails)
	contacts := emailsToContacts(c, r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), 0, nil
}

func GetSentEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("IsSent =", true).Filter("Cancel =", false).Filter("Delievered =", true).Filter("Archived =", false)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(c, r, emails)
	contacts := emailsToContacts(c, r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), 0, nil
}

func GetEmailStats(c context.Context, r *http.Request) (interface{}, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	timeseriesData, count, total, err := search.SearchEmailTimeseriesByUserId(c, r, user)
	return timeseriesData, nil, count, total, err
}

func GetScheduledEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	queryNoLimit := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	amountOfKeys, err := queryNoLimit.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(c, r, emails)
	contacts := emailsToContacts(c, r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), len(amountOfKeys), nil
}

func GetArchivedEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("Cancel =", false).Filter("IsSent =", true).Filter("Archived =", true)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(c, r, emails)
	contacts := emailsToContacts(c, r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), 0, nil
}

func GetTeamEmails(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	return []models.Email{}, nil, 0, 0, nil
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

func GetEmailByIdUnauthorized(c context.Context, r *http.Request, id int64) (models.Email, interface{}, error) {
	email, err := getEmailUnauthorized(c, r, id)
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
	includedContact := []models.Contact{}
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

	if email.ContactId != 0 {
		contact, err := getContact(c, r, email.ContactId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Email{}, nil, err
		}
		includedContact = append(includedContact, contact)
	}

	includes := make([]interface{}, len(includedFiles)+len(includedContact))

	for i := 0; i < len(includedFiles); i++ {
		includes[i] = includedFiles[i]
	}

	for i := 0; i < len(includedContact); i++ {
		includes[i+len(includedFiles)] = includedContact[i]
	}

	return email, includes, nil
}

/*
* Create methods
 */

func CreateEmailTransition(c context.Context, r *http.Request) ([]models.Email, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := controllers.GetCurrentUser(c, r)
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
			// Only valid if user is not using gmail, outlook, or smtp
			if emails[i].FromEmail != "" && !currentUser.Gmail && !currentUser.Outlook && !currentUser.ExternalEmail {
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

func UpdateEmail(c context.Context, r *http.Request, currentUser apiModels.User, email *models.Email, updatedEmail models.Email) (models.Email, interface{}, error) {
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

	user, err := controllers.GetCurrentUser(c, r)
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
	user, err := controllers.GetCurrentUser(c, r)
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

func CancelAllScheduled(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(c, ks, emails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
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
	return emails, nil, len(emails), 0, nil
}

func BulkCancelEmail(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var cancelEmails cancelEmailsBulk
	err := decoder.Decode(buf, &cancelEmails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails := []models.Email{}
	emailIds := []int64{} // Validated email ids
	for i := 0; i < len(cancelEmails.Emails); i++ {
		email, err := getEmail(c, r, cancelEmails.Emails[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			continue
		}

		// If it has not have a sentat date then we can cancel it
		// and that sendAt date is in the future.
		if !email.SendAt.IsZero() && email.SendAt.After(time.Now()) {
			email.Cancel = true
			email.Save(c)
			emails = append(emails, email)
			emailIds = append(emailIds, email.Id)
		}
	}

	sync.EmailResourceBulkSync(r, emailIds)
	return emails, nil, len(emails), 0, nil
}

func CancelEmail(c context.Context, r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	// If it has a sentat date then we can cancel it
	// and that sendAt date is in the future.
	if !email.SendAt.IsZero() && email.SendAt.After(time.Now()) {
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

func BulkSendEmail(c context.Context, r *http.Request) ([]models.Email, interface{}, int, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var bulkEmailIds models.BulkSendEmailIds
	err := decoder.Decode(buf, &bulkEmailIds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// If user is not active then they can't send
	// emails
	if !user.IsActive {
		return []models.Email{}, nil, 0, 0, err
	}

	if user.IsBanned {
		return []models.Email{}, nil, 0, 0, err
	}

	emails := []models.Email{}
	emailIds := []int64{}

	// Since the emails should be the same, get the attachments here
	if len(bulkEmailIds.EmailIds) > 0 {
		firstEmail, err := getEmail(c, r, bulkEmailIds.EmailIds[0])
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, 0, 0, err
		}

		files := []models.File{}
		if len(firstEmail.Attachments) > 0 {
			for i := 0; i < len(firstEmail.Attachments); i++ {
				file, err := getFile(c, r, firstEmail.Attachments[i])
				if err == nil {
					files = append(files, file)
				} else {
					log.Errorf(c, "%v", err)
				}
			}
		}

		bytesArray, attachmentType, fileNames, err := attach.GetAttachmentsForEmail(r, firstEmail, files)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		// Figure out what the emailMethod we should use
		emailMethod := "sendgrid"
		if user.SMTPValid && user.ExternalEmail && user.EmailSetting != 0 {
			emailMethod = "smtp"
		} else if user.AccessToken != "" && user.Gmail {
			emailMethod = "gmail"
		} else if user.OutlookAccessToken != "" && user.Outlook {
			emailMethod = "outlook"
		} else if user.UseSparkPost {
			emailMethod = "sparkpost"
		}

		emailSplit := 200
		if len(bulkEmailIds.EmailIds) < 200 {
			emailSplit = 40
		}

		betweenDelay := 150
		for i := 0; i < len(bulkEmailIds.EmailIds); i++ {
			delayAmount := int(float64(i) / float64(emailSplit))
			emailDelay := delayAmount * betweenDelay

			singleEmail, _, err := SendBulkEmailSingle(c, r, strconv.FormatInt(bulkEmailIds.EmailIds[i], 10), files, bytesArray, attachmentType, fileNames, emailDelay, emailMethod)
			if err != nil {
				log.Errorf(c, "%v", err)
			}
			emails = append(emails, singleEmail)
			emailIds = append(emailIds, singleEmail.Id)
		}

		sync.EmailResourceBulkSync(r, emailIds)
	}

	return emails, nil, len(emails), 0, nil
}

func SendEmail(c context.Context, r *http.Request, id string, isNotBulk bool) (models.Email, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	user, err := controllers.GetCurrentUser(c, r)
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

	if email.Subject == "" {
		email.Subject = "(no subject)"
	}

	email.Method = ""
	emailId := strconv.FormatInt(email.Id, 10)
	email.Body = utilities.AppendHrefWithLink(c, email.Body, emailId, "https://email2.newsai.co/a")
	email.Body += "<img src=\"https://email2.newsai.co/?id=" + emailId + "\" alt=\"NewsAI\" />"

	return email, nil, nil
}

func SendBulkEmailSingle(c context.Context, r *http.Request, id string, files []models.File, bytesArray [][]byte, attachmentType []string, fileNames []string, emailDelay int, method string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	user, err := controllers.GetCurrentUser(c, r)
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

	if email.Subject == "" {
		email.Subject = "(no subject)"
	}

	email.Method = method
	emailId := strconv.FormatInt(email.Id, 10)
	email.Body = utilities.AppendHrefWithLink(c, email.Body, emailId, "https://email2.newsai.co/a")
	email.Body += "<img src=\"https://email2.newsai.co/?id=" + emailId + "\" alt=\"NewsAI\" />"

	return email, nil, nil
}

func MarkBounced(c context.Context, r *http.Request, e *models.Email, reason string) (*models.Email, error) {
	controllers.SetUser(c, r, e.CreatedBy)

	contacts, err := filterContactByEmail(c, e.To)
	if err != nil {
		log.Infof(c, "%v", err)
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].EmailBounced = true
		contacts[i].Save(c, r)
	}

	_, err = e.MarkBounced(c, reason)
	return e, err
}

func MarkSpam(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(c, r, e.CreatedBy)
	_, err := e.MarkSpam(c)
	return e, err
}

func MarkClicked(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(c, r, e.CreatedBy)
	_, err := e.MarkClicked(c)
	return e, err
}

func MarkDelivered(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	_, err := e.MarkDelivered(c)
	return e, err
}

func MarkOpened(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(c, r, e.CreatedBy)
	_, err := e.MarkOpened(c)
	return e, err
}

func MarkSendgridOpen(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(c, r, e.CreatedBy)
	_, err := e.MarkSendgridOpened(c)
	return e, err
}

func MarkSendgridDrop(c context.Context, r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(c, r, e.CreatedBy)
	_, err := e.MarkSendgridDropped(c)
	return e, err
}

func GetEmailLogs(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	email, _, err := GetEmail(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Email{}, nil, err
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return email, nil, err
	}

	logs, _, _, err := search.SearchEmailLogByEmailId(c, r, user, email.Id)
	return logs, nil, err
}

func GetEmailSearch(c context.Context, r *http.Request) (interface{}, interface{}, int, int, error) {
	queryField := gcontext.Get(r, "q").(string)

	if queryField == "" {
		return nil, nil, 0, 0, nil
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	if strings.Contains(queryField, "date:") || strings.Contains(queryField, "subject:") || strings.Contains(queryField, "filter:") || strings.Contains(queryField, "baseSubject:") {
		emailFilters := strings.Split(queryField, ",")
		emailDate := ""
		emailSubject := ""
		emailBaseSubject := ""
		emailFilter := ""

		for i := 0; i < len(emailFilters); i++ {
			if strings.Contains(emailFilters[i], "date:") {
				emailDateArray := strings.Split(emailFilters[i], ":")
				if len(emailDateArray) > 1 {
					emailDate = strings.Join(emailDateArray[1:], ":")
					emailDate = strings.Replace(emailDate, "\\", "", -1)

					if last := len(emailDate) - 1; last >= 0 && emailDate[last] == '"' {
						emailDate = emailDate[:last]
					}

					if emailDate[0] == '"' {
						emailDate = emailDate[1:]
					}
				}
			} else if strings.Contains(emailFilters[i], "filter:") {
				emailFilterArray := strings.Split(emailFilters[i], ":")
				if len(emailFilterArray) > 1 {
					emailFilter = strings.Join(emailFilterArray[1:], ":")
					emailFilter = strings.Replace(emailFilter, "\\", "", -1)

					if last := len(emailFilter) - 1; last >= 0 && emailFilter[last] == '"' {
						emailFilter = emailFilter[:last]
					}

					if emailFilter[0] == '"' {
						emailFilter = emailFilter[1:]
					}
				}
			} else if strings.Contains(emailFilters[i], "subject:") {
				if len(emailFilters) > 2 {
					emailSubjectSplit := strings.Split(queryField, "subject:")
					emailFilters[i] = "subject:" + emailSubjectSplit[len(emailSubjectSplit)-1]
				}
				emailSubjectArray := strings.Split(emailFilters[i], ":")
				if len(emailSubjectArray) > 1 {
					log.Infof(c, "%v", emailSubjectArray)
					// Recover the pieces when split by colon
					emailSubject = strings.Join(emailSubjectArray[1:], ":")
					emailSubject = strings.Replace(emailSubject, "\\", "", -1)

					if last := len(emailSubject) - 1; last >= 0 && emailSubject[last] == '"' {
						emailSubject = emailSubject[:last]
					}

					if emailSubject[0] == '"' {
						emailSubject = emailSubject[1:]
					}

					log.Infof(c, "%v", emailSubject)
				}
			} else if strings.Contains(emailFilters[i], "baseSubject:") {
				emailBaseSubjectArray := strings.Split(emailFilters[i], ":")
				if len(emailBaseSubjectArray) > 1 {
					// Recover the pieces when split by colon
					emailBaseSubject = strings.Join(emailBaseSubjectArray[1:], ":")
					emailBaseSubject = strings.Replace(emailBaseSubject, "\\", "", -1)

					if last := len(emailBaseSubject) - 1; last >= 0 && emailBaseSubject[last] == '"' {
						emailBaseSubject = emailBaseSubject[:last]
					}

					if emailBaseSubject[0] == '"' {
						emailBaseSubject = emailBaseSubject[1:]
					}
				}
			}
		}

		if emailDate != "" || emailSubject != "" || emailFilter != "" || emailBaseSubject != "" {
			emails, count, total, err := search.SearchEmailsByQueryFields(c, r, user, emailDate, emailSubject, emailBaseSubject, emailFilter)

			// Add includes
			mediaLists := emailsToLists(c, r, emails)
			contacts := emailsToContacts(c, r, emails)
			includes := make([]interface{}, len(mediaLists)+len(contacts))
			for i := 0; i < len(mediaLists); i++ {
				includes[i] = mediaLists[i]
			}

			for i := 0; i < len(contacts); i++ {
				includes[i+len(mediaLists)] = contacts[i]
			}

			return emails, includes, count, total, err
		} else {
			return nil, nil, 0, 0, errors.New("Please enter a valid date or subject")
		}
	}

	emails, count, total, err := search.SearchEmailsByQuery(c, r, user, queryField)

	// Add includes
	mediaLists := emailsToLists(c, r, emails)
	contacts := emailsToContacts(c, r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, count, total, err
}

func GetEmailCampaigns(c context.Context, r *http.Request) (interface{}, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	emails, count, total, err := search.SearchEmailCampaignsByDate(c, r, user)
	return emails, nil, count, total, err
}

func GetEmailCampaignsForUser(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	user := apiModels.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = controllers.GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, 0, 0, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, 0, 0, err
		}
		user, _, err = controllers.GetUserById(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Email{}, nil, 0, 0, err
		}
	}

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails, count, total, err := search.SearchEmailCampaignsByDate(c, r, user)
	return emails, nil, count, total, err
}

func GetEmailProviderLimits(c context.Context, r *http.Request) (interface{}, interface{}, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	emailProviderLimits := models.EmailProviderLimits{}
	emailProviderLimits.SendGridLimits = 2000
	emailProviderLimits.OutlookLimits = 500
	emailProviderLimits.GmailLimits = 500
	emailProviderLimits.SMTPLimits = 2000

	t := time.Now()
	todayDateMorning := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	todayDateNight := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 59, time.Local)

	// SendGrid
	sendGrid, err := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("Method =", "sendgrid").Filter("IsSent =", true).Filter("Delievered =", true).Filter("Created <=", todayDateNight).Filter("Created >=", todayDateMorning).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}
	emailProviderLimits.SendGrid = len(sendGrid)

	// Outlook

	return emailProviderLimits, nil, nil
}
