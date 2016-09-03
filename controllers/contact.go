package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getContact(c context.Context, r *http.Request, id int64) (models.Contact, error) {
	if id == 0 {
		return models.Contact{}, errors.New("datastore: no such entity")
	}
	// Get the Contact by id
	var contact models.Contact
	contactId := datastore.NewKey(c, "Contact", "", id, nil)
	err := nds.Get(c, contactId, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	if !contact.Created.IsZero() {
		contact.Format(contactId, "contacts")

		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		if !contact.IsMasterContact && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		return contact, nil
	}
	return models.Contact{}, errors.New("No contact by this id")
}

/*
* Update methods
 */

func updateContact(c context.Context, r *http.Request, contact *models.Contact, updatedContact models.Contact) (models.Contact, interface{}, error) {
	utilities.UpdateIfNotBlank(&contact.FirstName, updatedContact.FirstName)
	utilities.UpdateIfNotBlank(&contact.LastName, updatedContact.LastName)
	utilities.UpdateIfNotBlank(&contact.Email, updatedContact.Email)
	utilities.UpdateIfNotBlank(&contact.LinkedIn, updatedContact.LinkedIn)
	utilities.UpdateIfNotBlank(&contact.Twitter, updatedContact.Twitter)
	utilities.UpdateIfNotBlank(&contact.Instagram, updatedContact.Instagram)
	utilities.UpdateIfNotBlank(&contact.Website, updatedContact.Website)
	utilities.UpdateIfNotBlank(&contact.Blog, updatedContact.Blog)
	utilities.UpdateIfNotBlank(&contact.Notes, updatedContact.Notes)

	if len(updatedContact.CustomFields) > 0 {
		contact.CustomFields = updatedContact.CustomFields
	}

	if len(updatedContact.Employers) > 0 {
		contact.Employers = updatedContact.Employers
	}

	if len(updatedContact.PastEmployers) > 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	// Special case when you want to remove all the employers
	if len(contact.Employers) > 0 && len(updatedContact.Employers) == 0 {
		contact.Employers = updatedContact.Employers
	}

	// Special case when you want to remove all the past employers
	if len(contact.PastEmployers) > 0 && len(updatedContact.PastEmployers) == 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	_, err := Save(c, r, contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	// Logging the action happening
	LogNotificationForResource(c, r, "Contact", contact.Id, "UPDATE", "")

	return *contact, nil, nil
}

func updateSocial(c context.Context, r *http.Request, contact *models.Contact, updatedContact models.Contact) (models.Contact, interface{}, error) {
	utilities.UpdateIfNotBlank(&contact.LinkedIn, updatedContact.LinkedIn)
	utilities.UpdateIfNotBlank(&contact.Twitter, updatedContact.Twitter)
	utilities.UpdateIfNotBlank(&contact.Instagram, updatedContact.Instagram)
	utilities.UpdateIfNotBlank(&contact.Website, updatedContact.Website)
	utilities.UpdateIfNotBlank(&contact.Blog, updatedContact.Blog)

	_, err := Save(c, r, contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	// Logging the action happening
	LogNotificationForResource(c, r, "Contact", contact.Id, "UPDATE", "SOCIAL")

	return *contact, nil, nil
}

/*
* Filter methods
 */

func filterContact(c context.Context, r *http.Request, queryType, query string) (models.Contact, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	if len(contacts) > 0 {
		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		if !contacts[0].IsMasterContact && !permissions.AccessToObject(contacts[0].CreatedBy, user.Id) {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}
		contacts[0].Format(ks[0], "contacts")
		return contacts[0], nil
	}
	return models.Contact{}, errors.New("No contact by this " + queryType)
}

/*
* Normalization methods
 */

func checkAgainstParent(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	// If there is a parent contact
	if ct.ParentContact != 0 {
		// Get parent contact
		parentContact, err := getContact(c, r, ct.ParentContact)
		if err != nil {
			log.Errorf(c, "%v", err)
			return ct, err
		}

		// See differences in parent and child contact
		if !reflect.DeepEqual(ct.Employers, parentContact.Employers) || !reflect.DeepEqual(ct.PastEmployers, parentContact.PastEmployers) {
			ct.IsOutdated = true
			Save(c, r, ct)
		}

		return ct, nil
	}
	return ct, nil
}

func socialSync(c context.Context, r *http.Request, ct *models.Contact, justCreated bool) (*models.Contact, error) {
	if ct.ParentContact == 0 {
		return ct, nil
	}

	parentContact, err := getContact(c, r, ct.ParentContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return ct, err
	}

	hourFromUpdate := parentContact.LinkedInUpdated.Add(time.Hour * 1)

	// Update LinkedIn contact
	if parentContact.IsMasterContact && parentContact.LinkedIn != "" && (time.Now().After(hourFromUpdate) || parentContact.LinkedInUpdated.IsZero()) {
		// Send a pub to Influencer
		err = sync.SocialSync(r, "linkedinUrl", parentContact.LinkedIn, parentContact.Id, justCreated)

		if err != nil {
			log.Errorf(c, "%v", err)
			return ct, err
		}

		LogNotificationForResource(c, r, "Contact", ct.Id, "SYNC", "LINKEDIN")
		LogNotificationForResource(c, r, "Contact", parentContact.Id, "SYNC", "LINKEDIN")

		// Now that we have told the Influencer program that we are syncing Linkedin data
		parentContact.LinkedInUpdated = time.Now()
		parentContact.Save(c, r)
	}

	return ct, nil
}

func filterMasterContact(c context.Context, r *http.Request, ct *models.Contact, queryType, query string) (models.Contact, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" = ", query).Filter("IsMasterContact = ", true).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	if len(ks) == 0 {
		return models.Contact{}, errors.New("No contact by the field " + queryType)
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	if len(contacts) > 0 {
		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, errors.New("Could not get user")
		}

		if !contacts[0].IsMasterContact && !permissions.AccessToObject(contacts[0].CreatedBy, user.Id) {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		contacts[0].Format(ks[0], "contacts")
		return contacts[0], nil
	}
	return models.Contact{}, errors.New("No contact by this " + queryType)
}

func findOrCreateMasterContact(c context.Context, ct *models.Contact, r *http.Request) (*models.Contact, error, bool) {
	// Find master contact
	// If there is no parent contact Id or if the Linkedin field is not empty
	if ct.ParentContact == 0 && ct.LinkedIn != "" {
		masterContact, err := filterMasterContact(c, r, ct, "LinkedIn", ct.LinkedIn)
		// Master contact does not exist
		if err != nil {
			// Create master contact
			newMasterContact := models.Contact{}

			// Initialize with the same values
			newMasterContact.FirstName = ct.FirstName
			newMasterContact.LastName = ct.LastName
			newMasterContact.Email = ct.Email
			newMasterContact.Employers = ct.Employers
			newMasterContact.LinkedIn = ct.LinkedIn
			newMasterContact.Twitter = ct.Twitter
			newMasterContact.Instagram = ct.Instagram
			newMasterContact.MuckRack = ct.MuckRack
			newMasterContact.Website = ct.Website
			newMasterContact.Blog = ct.Blog

			// Set this to be the new master contact
			newMasterContact.IsMasterContact = true

			// Create the new master contact
			Create(c, r, &newMasterContact)

			// Do a social sync task when new master contact is added

			// Assign the Id of the parent contact to be the new master contact.
			ct.ParentContact = newMasterContact.Id
			ct.IsMasterContact = false

			// Logging the action happening
			LogNotificationForResource(c, r, "Contact", ct.Id, "CREATE", "PARENT")

			return ct, nil, true
		}

		// Update social information

		// Don't create master contact
		ct.ParentContact = masterContact.Id
		return ct, nil, false
	}

	return ct, nil, false
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single contact
func GetContacts(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	queryField := gcontext.Get(r, "query").(string)
	if queryField != "" {
		contacts, err := search.SearchContact(c, r, queryField, user.Id, 0)
		if err != nil {
			return []models.Contact{}, nil, 0, err
		}
		return contacts, nil, len(contacts), nil
	}

	query := datastore.NewQuery("Contact").Filter("CreatedBy =", user.Id).Filter("IsMasterContact = ", false)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	contacts := []models.Contact{}
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contacts, nil, 0, err
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].Format(ks[i], "contacts")
	}

	return contacts, nil, len(contacts), nil
}

func GetContact(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	if contact.LinkedIn != "" {
		_, err = socialSync(c, r, &contact, false)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		checkAgainstParent(c, r, &contact)
	}

	return contact, nil, nil
}

func GetDiff(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	// Get parent contact
	parentContactId := strconv.FormatInt(contact.ParentContact, 10)
	parentContact, _, err := GetContact(c, r, parentContactId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	newEmployers := []string{}
	newPastEmployers := []string{}
	for i := 0; i < len(parentContact.Employers); i++ {
		// Get each publication
		currentPublication, err := getPublication(c, parentContact.Employers[i])
		if err != nil {
			err = errors.New("Only actions are diff and update")
			log.Errorf(c, "%v", err)
			return nil, nil, err
		}
		newEmployers = append(newEmployers, currentPublication.Name)
	}

	for i := 0; i < len(parentContact.PastEmployers); i++ {
		// Get each publication
		currentPublication, err := getPublication(c, parentContact.PastEmployers[i])
		if err != nil {
			err = errors.New("Only actions are diff and update")
			log.Errorf(c, "%v", err)
			return nil, nil, err
		}
		newPastEmployers = append(newPastEmployers, currentPublication.Name)
	}

	data := struct {
		NewEmployers     []string `json:"employers"`
		NewPastEmployers []string `json:"pastemployers"`
	}{
		newEmployers,
		newPastEmployers,
	}

	return data, nil, nil
}

/*
* Create methods
 */

func Create(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return ct, err
	}

	ct.Create(c, r, currentUser)

	// Logging the action happening
	LogNotificationForResource(c, r, "Contact", ct.Id, "CREATE", "")

	if ct.ParentContact == 0 && !ct.IsMasterContact {
		_, _, justCreated := findOrCreateMasterContact(c, ct, r)
		socialSync(c, r, ct, justCreated)
		checkAgainstParent(c, r, ct)
	}

	_, err = Save(c, r, ct)
	return ct, err
}

func CreateContact(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	decoder := ffjson.NewDecoder()
	var contact models.Contact
	err := decoder.Decode(buf, &contact)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var contacts []models.Contact

		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &contacts)

		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		newContacts := []models.Contact{}
		for i := 0; i < len(contacts); i++ {
			_, err = Create(c, r, &contacts[i])
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Contact{}, nil, 0, err
			}
			newContacts = append(newContacts, contacts[i])
		}

		return newContacts, nil, len(newContacts), nil
	}

	// Create contact
	_, err = Create(c, r, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	return []models.Contact{contact}, nil, 0, nil
}

func BatchCreateContactsForExcelUpload(c context.Context, r *http.Request, contacts []models.Contact) ([]int64, error) {
	var keys []*datastore.Key
	var contactIds []int64

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, err
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].CreatedBy = currentUser.Id
		contacts[i].Created = time.Now()
		contacts[i].Updated = time.Now()
		contacts[i].Normalize()
		keys = append(keys, contacts[i].Key(c))

		if contacts[i].ParentContact == 0 && !contacts[i].IsMasterContact && contacts[i].LinkedIn != "" {
			findOrCreateMasterContact(c, &contacts[i], r)
			socialSync(c, r, &contacts[i], false)
			checkAgainstParent(c, r, &contacts[i])
		}
	}

	ks := []*datastore.Key{}

	err = nds.RunInTransaction(c, func(ctx context.Context) error {
		contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
		ks, err = nds.PutMulti(contextWithTimeout, keys, contacts)
		if err != nil {
			log.Errorf(c, "%v", err)
			return err
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, err
	}

	for i := 0; i < len(ks); i++ {
		contactIds = append(contactIds, ks[i].IntID())
	}

	return contactIds, nil
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func Save(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	// Update the Updated time
	ct.Normalize()

	if ct.ParentContact == 0 && !ct.IsMasterContact {
		findOrCreateMasterContact(c, ct, r)
		socialSync(c, r, ct, false)
		checkAgainstParent(c, r, ct)
	}

	ct.Save(c, r)

	return ct, nil
}

func UpdateSingleContact(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get the details of the current contact
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, errors.New("Could not get user")
	}

	if !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
		return models.Contact{}, nil, errors.New("You don't have permissions to edit these objects")
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedContact models.Contact
	err = decoder.Decode(buf, &updatedContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	return updateContact(c, r, &contact, updatedContact)
}

func UpdateBatchContact(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedContacts []models.Contact
	err := decoder.Decode(buf, &updatedContacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	// Get logged in user
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, errors.New("Could not get user")
	}

	// Check if each of the contacts have permissions before updating anything
	currentContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		contact, err := getContact(c, r, updatedContacts[i].Id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		if !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			return []models.Contact{}, nil, 0, errors.New("Forbidden")
		}

		currentContacts = append(currentContacts, contact)
	}

	// Update each of the contacts
	newContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		updatedContact, _, err := updateContact(c, r, &currentContacts[i], updatedContacts[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		newContacts = append(newContacts, updatedContact)
	}

	return newContacts, nil, len(newContacts), nil
}

/*
* Normalization methods
 */

func UpdateContactToParent(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get current contact
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contact, nil, err
	}

	if !contact.IsOutdated {
		return contact, nil, nil
	}

	parentContact, err := getContact(c, r, contact.ParentContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contact, nil, err
	}

	contact.Employers = parentContact.Employers
	contact.PastEmployers = parentContact.PastEmployers
	contact.IsOutdated = false
	_, err = Save(c, r, &contact)

	// Logging the action happening
	LogNotificationForResource(c, r, "Contact", contact.Id, "UPDATE", "TO_PARENT")

	if err != nil {
		log.Errorf(c, "%v", err)
		return contact, nil, err
	}

	return contact, nil, nil
}

func SocialSync(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contact, nil, err
	}

	if contact.ParentContact == 0 {
		return contact, nil, nil
	}

	parentContact, err := getContact(c, r, contact.ParentContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contact, nil, err
	}

	// Send a pub to Influencer
	err = sync.SocialSync(r, "linkedinUrl", parentContact.LinkedIn, parentContact.Id, false)

	if err != nil {
		log.Errorf(c, "%v", err)
		return contact, nil, err
	}

	// Now that we have told the Influencer program that we are syncing Linkedin data
	parentContact.LinkedInUpdated = time.Now()
	parentContact.Save(c, r)

	// Logging the action happening
	LogNotificationForResource(c, r, "Contact", contact.Id, "SYNC", "LINKEDIN")
	LogNotificationForResource(c, r, "Contact", parentContact.Id, "SYNC", "LINKEDIN")

	return contact, nil, nil
}
