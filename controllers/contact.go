package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/permissions"
	"github.com/news-ai/tabulae/sync"
	"github.com/news-ai/tabulae/utils"
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
		return models.Contact{}, err
	}

	if !contact.Created.IsZero() {
		contact.Id = contactId.IntID()

		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.Contact{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			return models.Contact{}, errors.New("Forbidden")
		}

		return contact, nil
	}
	return models.Contact{}, errors.New("No contact by this id")
}

/*
* Filter methods
 */

func filterContact(c context.Context, r *http.Request, queryType, query string) (models.Contact, error) {
	// Get an contact by a query type
	contacts := []models.Contact{}
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).GetAll(c, &contacts)
	if err != nil {
		return models.Contact{}, err
	}
	if len(contacts) > 0 {
		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.Contact{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(contacts[0].CreatedBy, user.Id) {
			return models.Contact{}, errors.New("Forbidden")
		}

		contacts[0].Id = ks[0].IntID()
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

func linkedInSync(c context.Context, r *http.Request, ct *models.Contact, justCreated bool) (*models.Contact, error) {
	if ct.ParentContact == 0 {
		return ct, nil
	}

	parentContact, err := getContact(c, r, ct.ParentContact)
	if err != nil {
		return ct, err
	}

	hourFromUpdate := parentContact.LinkedInUpdated.Add(time.Hour * 1)

	// Update LinkedIn contact
	if parentContact.IsMasterContact && parentContact.LinkedIn != "" && (time.Now().After(hourFromUpdate) || parentContact.LinkedInUpdated.IsZero()) {
		// Send a pub to Influencer
		err = sync.LinkedInSync(r, parentContact.LinkedIn, parentContact.Id, justCreated)

		if err != nil {
			log.Errorf(c, "%v", err)
			return ct, err
		}

		// Now that we have told the Influencer program that we are syncing Linkedin data
		parentContact.LinkedInUpdated = time.Now()
		parentContact.Save(c, r)
	}

	return ct, nil
}

func filterMasterContact(c context.Context, r *http.Request, ct *models.Contact, queryType, query string) (models.Contact, error) {
	// Get an contact by a query type
	contacts := []models.Contact{}
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).Filter("IsMasterContact = ", true).GetAll(c, &contacts)
	if err != nil {
		return models.Contact{}, err
	}
	if len(contacts) > 0 {
		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.Contact{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(contacts[0].CreatedBy, user.Id) {
			return models.Contact{}, errors.New("Forbidden")
		}

		contacts[0].Id = ks[0].IntID()
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
			return ct, nil, true
		}

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
func GetContacts(c context.Context, r *http.Request) ([]models.Contact, error) {
	contacts := []models.Contact{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Contact{}, err
	}

	ks, err := datastore.NewQuery("Contact").Filter("CreatedBy =", user.Id).Filter("IsMasterContact = ", false).GetAll(c, &contacts)
	if err != nil {
		return []models.Contact{}, err
	}
	for i := 0; i < len(contacts); i++ {
		contacts[i].Id = ks[i].IntID()
	}

	return contacts, nil
}

func GetContact(c context.Context, r *http.Request, id string) (models.Contact, error) {
	// Get the details of the current user
	currentId, err := utils.StringIdToInt(id)
	if err != nil {
		return models.Contact{}, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		return models.Contact{}, err
	}

	if contact.LinkedIn != "" {
		_, err = linkedInSync(c, r, &contact, false)
		if err != nil {
			log.Errorf(c, "%v", err.Error())
		}

		checkAgainstParent(c, r, &contact)
	}

	return contact, nil
}

/*
* Create methods
 */

func Create(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return ct, err
	}

	ct.Create(c, r, currentUser)

	if ct.ParentContact == 0 && !ct.IsMasterContact {
		_, _, justCreated := findOrCreateMasterContact(c, ct, r)
		linkedInSync(c, r, ct, justCreated)
		checkAgainstParent(c, r, ct)
	}

	_, err = Save(c, r, ct)
	return ct, err
}

func CreateContact(c context.Context, r *http.Request) ([]models.Contact, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))

	decoder := json.NewDecoder(rdr1)
	var contact models.Contact
	err := decoder.Decode(&contact)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var contacts []models.Contact

		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
		arrayDecoder := json.NewDecoder(rdr2)
		err = arrayDecoder.Decode(&contacts)

		if err != nil {
			return []models.Contact{}, err
		}

		newContacts := []models.Contact{}
		for i := 0; i < len(contacts); i++ {
			_, err = Create(c, r, &contacts[i])
			if err != nil {
				return []models.Contact{}, err
			}
			newContacts = append(newContacts, contacts[i])
		}

		return newContacts, nil
	}

	// Create contact
	_, err = Create(c, r, &contact)
	if err != nil {
		return []models.Contact{}, err
	}

	return []models.Contact{contact}, nil
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
		linkedInSync(c, r, ct, false)
		checkAgainstParent(c, r, ct)
	}

	ct.Save(c, r)

	return ct, nil
}

func UpdateContact(c context.Context, r *http.Request, contact *models.Contact, updatedContact models.Contact) (models.Contact, error) {
	utils.UpdateIfNotBlank(&contact.FirstName, updatedContact.FirstName)
	utils.UpdateIfNotBlank(&contact.LastName, updatedContact.LastName)
	utils.UpdateIfNotBlank(&contact.Email, updatedContact.Email)
	utils.UpdateIfNotBlank(&contact.LinkedIn, updatedContact.LinkedIn)
	utils.UpdateIfNotBlank(&contact.Twitter, updatedContact.Twitter)
	utils.UpdateIfNotBlank(&contact.Instagram, updatedContact.Instagram)
	utils.UpdateIfNotBlank(&contact.Website, updatedContact.Website)
	utils.UpdateIfNotBlank(&contact.Blog, updatedContact.Blog)
	utils.UpdateIfNotBlank(&contact.Notes, updatedContact.Notes)

	if len(updatedContact.CustomFields) > 0 {
		contact.CustomFields = updatedContact.CustomFields
	}

	if len(updatedContact.Employers) > 0 {
		contact.Employers = updatedContact.Employers
	}

	if len(updatedContact.PastEmployers) > 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	Save(c, r, contact)

	return *contact, nil
}

func UpdateSingleContact(c context.Context, r *http.Request, id string) (models.Contact, error) {
	// Get the details of the current contact
	contact, err := GetContact(c, r, id)
	if err != nil {
		return models.Contact{}, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return models.Contact{}, errors.New("Could not get user")
	}

	if !permissions.AccessToObject(contact.CreatedBy, user.Id) {
		return models.Contact{}, errors.New("You don't have permissions to edit these objects")
	}

	decoder := json.NewDecoder(r.Body)
	var updatedContact models.Contact
	err = decoder.Decode(&updatedContact)
	if err != nil {
		return models.Contact{}, err
	}

	return UpdateContact(c, r, &contact, updatedContact)
}

func UpdateBatchContact(c context.Context, r *http.Request) ([]models.Contact, error) {
	decoder := json.NewDecoder(r.Body)
	var updatedContacts []models.Contact
	err := decoder.Decode(&updatedContacts)
	if err != nil {
		return []models.Contact{}, err
	}

	// Get logged in user
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Contact{}, errors.New("Could not get user")
	}

	// Check if each of the contacts have permissions before updating anything
	currentContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		contact, err := getContact(c, r, updatedContacts[i].Id)
		if err != nil {
			return []models.Contact{}, err
		}

		if !permissions.AccessToObject(contact.CreatedBy, user.Id) {
			return []models.Contact{}, errors.New("Forbidden")
		}

		currentContacts = append(currentContacts, contact)
	}

	// Update each of the contacts
	newContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		updatedContact, err := UpdateContact(c, r, &currentContacts[i], updatedContacts[i])
		if err != nil {
			return []models.Contact{}, err
		}

		newContacts = append(newContacts, updatedContact)
	}

	return newContacts, nil
}

/*
* Normalization methods
 */

func UpdateContactToParent(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	parentContact, err := getContact(c, r, ct.ParentContact)

	if err != nil {
		return ct, err
	}

	ct.Employers = parentContact.Employers
	ct.IsOutdated = false
	_, err = Save(c, r, ct)

	if err != nil {
		return ct, err
	}

	return ct, nil
}

func LinkedInSync(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	if ct.ParentContact == 0 {
		return ct, nil
	}

	parentContact, err := getContact(c, r, ct.ParentContact)
	if err != nil {
		return ct, err
	}

	// Send a pub to Influencer
	err = sync.LinkedInSync(r, parentContact.LinkedIn, parentContact.Id, false)

	if err != nil {
		log.Errorf(c, "%v", err)
		return ct, err
	}

	// Now that we have told the Influencer program that we are syncing Linkedin data
	parentContact.LinkedInUpdated = time.Now()
	parentContact.Save(c, r)
	return ct, nil
}
