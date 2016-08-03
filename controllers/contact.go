package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/sync"
)

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (ct *Contact) Key(c appengine.Context) *datastore.Key {
	if ct.Id == 0 {
		ct.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Contact", nil)
	}
	return datastore.NewKey(c, "Contact", "", ct.Id, nil)
}

/*
* Get methods
 */

func getContact(c appengine.Context, r *http.Request, id int64) (models.Contact, error) {
	// Get the Contact by id
	contacts := []models.Contact{}
	contactId := datastore.NewKey(c, "Contact", "", id, nil)
	ks, err := datastore.NewQuery("Contact").Filter("__key__ =", contactId).GetAll(c, &contacts)
	if err != nil {
		return models.Contact{}, err
	}
	if len(contacts) > 0 {
		contacts[0].Id = ks[0].IntID()

		// If there is a parent
		if contacts[0].ParentContact != 0 {
			// Update information
			// contacts[0].linkedInSync(c, r)
			contacts[0].checkAgainstParent(c, r)
		}

		return contacts[0], nil
	}
	return models.Contact{}, errors.New("No contact by this id")
}

/*
* Create methods
 */

func (ct *Contact) Create(c appengine.Context, r *http.Request) (*Contact, error) {
	currentUser, err := getCurrentUser(c, r)
	if err != nil {
		return ct, err
	}

	ct.CreatedBy = currentUser.Id
	ct.Created = time.Now()
	ct.noramlize()

	if ct.ParentContact == 0 && !ct.IsMasterContact {
		ct.findOrCreateMasterContact(c, r)
		// ct.linkedInSync(c, r)
		ct.checkAgainstParent(c, r)
	}

	_, err = ct.Save(c, r)
	return ct, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ct *Contact) Save(c appengine.Context, r *http.Request) (*Contact, error) {
	// Update the Updated time
	ct.Updated = time.Now()
	ct.noramlize()

	if ct.ParentContact == 0 && !ct.IsMasterContact {
		ct.findOrCreateMasterContact(c, r)
		// ct.linkedInSync(c, r)
		ct.checkAgainstParent(c, r)
	}

	k, err := datastore.Put(c, ct.Key(c), ct)
	if err != nil {
		return nil, err
	}
	ct.Id = k.IntID()
	return ct, nil
}

/*
* Filter methods
 */

func filterContact(c appengine.Context, queryType, query string) (models.Contact, error) {
	// Get an contact by a query type
	contacts := []models.Contact{}
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).GetAll(c, &contacts)
	if err != nil {
		return models.Contact{}, err
	}
	if len(contacts) > 0 {
		contacts[0].Id = ks[0].IntID()
		return contacts[0], nil
	}
	return models.Contact{}, errors.New("No contact by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single contact
func GetContacts(c appengine.Context, r *http.Request) ([]models.Contact, error) {
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

func GetContact(c appengine.Context, r *http.Request, id string) (models.Contact, error) {
	// Get the details of the current user
	currentId, err := StringIdToInt(id)
	if err != nil {
		return models.Contact{}, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		return models.Contact{}, err
	}

	return contact, nil
}

/*
* Create methods
 */

func CreateContact(c appengine.Context, r *http.Request) ([]models.Contact, error) {
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
			_, err = contacts[i].Create(c, r)
			if err != nil {
				return []models.Contact{}, err
			}
			newContacts = append(newContacts, contacts[i])
		}

		return newContacts, nil
	}

	// Create contact
	_, err = contact.Create(c, r)
	if err != nil {
		return []models.Contact{}, err
	}

	return []models.Contact{contact}, nil
}

/*
* Update methods
 */

func UpdateContact(c appengine.Context, r *http.Request, contact *models.Contact, updatedContact models.Contact) models.Contact {
	UpdateIfNotBlank(&contact.FirstName, updatedContact.FirstName)
	UpdateIfNotBlank(&contact.LastName, updatedContact.LastName)
	UpdateIfNotBlank(&contact.Email, updatedContact.Email)
	UpdateIfNotBlank(&contact.LinkedIn, updatedContact.LinkedIn)
	UpdateIfNotBlank(&contact.Twitter, updatedContact.Twitter)
	UpdateIfNotBlank(&contact.Instagram, updatedContact.Instagram)
	UpdateIfNotBlank(&contact.Website, updatedContact.Website)
	UpdateIfNotBlank(&contact.Blog, updatedContact.Blog)
	UpdateIfNotBlank(&contact.Notes, updatedContact.Notes)

	if len(updatedContact.CustomFields) > 0 {
		contact.CustomFields = updatedContact.CustomFields
	}

	if len(updatedContact.Employers) > 0 {
		contact.Employers = updatedContact.Employers
	}

	if len(updatedContact.PastEmployers) > 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	contact.Save(c, r)

	return *contact
}

func UpdateSingleContact(c appengine.Context, r *http.Request, id string) (models.Contact, error) {
	// Get the details of the current contact
	contact, err := GetContact(c, r, id)
	if err != nil {
		return models.Contact{}, err
	}

	decoder := json.NewDecoder(r.Body)
	var updatedContact models.Contact
	err = decoder.Decode(&updatedContact)
	if err != nil {
		return models.Contact{}, err
	}

	return UpdateContact(c, r, &contact, updatedContact), nil
}

func UpdateBatchContact(c appengine.Context, r *http.Request) ([]models.Contact, error) {
	decoder := json.NewDecoder(r.Body)
	var updatedContacts []models.Contact
	err := decoder.Decode(&updatedContacts)
	if err != nil {
		return []models.Contact{}, err
	}

	newContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		contact, err := getContact(c, r, updatedContacts[i].Id)
		if err != nil {
			return []models.Contact{}, err
		}
		updatedContact := UpdateContact(c, r, &contact, updatedContacts[i])
		newContacts = append(newContacts, updatedContact)
	}

	return newContacts, nil
}

/*
* Normalization methods
 */

func (ct *Contact) noramlize() (*Contact, error) {
	ct.LinkedIn = stripQueryString(ct.LinkedIn)
	ct.Twitter = stripQueryString(ct.Twitter)
	ct.Instagram = stripQueryString(ct.Instagram)
	ct.MuckRack = stripQueryString(ct.MuckRack)
	ct.Website = stripQueryString(ct.Website)
	ct.Blog = stripQueryString(ct.Blog)

	return ct, nil
}

func (ct *Contact) checkAgainstParent(c appengine.Context, r *http.Request) (*Contact, error) {
	// If there is a parent contact
	if ct.ParentContact != 0 {
		// Get parent contact
		parentContact, err := getContact(c, r, ct.ParentContact)
		if err != nil {
			return ct, err
		}

		// See differences in parent and child contact
		if !reflect.DeepEqual(ct.Employers, parentContact.Employers) {
			ct.IsOutdated = true
			ct.Save(c, r)
		}

		return ct, nil
	}
	return ct, nil
}

func (ct *Contact) linkedInSync(c appengine.Context, r *http.Request) (*Contact, error) {
	parentContact, err := getContact(c, r, ct.ParentContact)
	if err != nil {
		return ct, err
	}
	// Update LinkedIn contact
	hourFromUpdate := parentContact.LinkedInUpdated.Add(time.Minute * 1)

	if parentContact.IsMasterContact && parentContact.LinkedIn != "" && (!parentContact.LinkedInUpdated.Before(hourFromUpdate) || parentContact.LinkedInUpdated.IsZero()) {
		linkedInData := sync.LinkedInSync(r, parentContact.LinkedIn)
		newEmployers := []int64{}
		// Update data through linkedin data
		for i := 0; i < len(linkedInData.Current); i++ {
			employerName := linkedInData.Current[i].Employer
			employer, err := FindOrCreatePublication(c, r, employerName)
			if err == nil {
				newEmployers = append(newEmployers, employer.Id)
			}
		}

		parentContact.Employers = newEmployers
		parentContact.LinkedInUpdated = time.Now()
		parentContact.Save(c, r)

		ct.LinkedInUpdated = time.Now()
		ct.Save(c, r)
	}

	return ct, nil
}

func (ct *Contact) UpdateContactToParent(c appengine.Context, r *http.Request) (*Contact, error) {
	parentContact, err := getContact(c, r, ct.ParentContact)

	if err != nil {
		return ct, err
	}

	ct.Employers = parentContact.Employers
	ct.IsOutdated = false
	_, err = ct.Save(c, r)

	if err != nil {
		return ct, err
	}

	return ct, nil
}

func (ct *Contact) filterMasterContact(c appengine.Context, queryType, query string) (Contact, error) {
	// Get an contact by a query type
	contacts := []Contact{}
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).Filter("IsMasterContact = ", true).GetAll(c, &contacts)
	if err != nil {
		return Contact{}, err
	}
	if len(contacts) > 0 {
		contacts[0].Id = ks[0].IntID()
		return contacts[0], nil
	}
	return Contact{}, errors.New("No contact by this " + queryType)
}

func (ct *Contact) findOrCreateMasterContact(c appengine.Context, r *http.Request) (*Contact, error) {
	// Find master contact
	// If there is no parent contact Id or if the Linkedin field is not empty
	if ct.ParentContact == 0 && ct.LinkedIn != "" {
		masterContact, err := ct.filterMasterContact(c, "LinkedIn", ct.LinkedIn)
		// Master contact does not exist
		if err != nil {
			// Create master contact
			newMasterContact := Contact{}

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
			newMasterContact.Create(c, r)

			// Do a social sync task when new master contact is added

			// Assign the Id of the parent contact to be the new master contact.
			ct.ParentContact = newMasterContact.Id
			ct.IsMasterContact = false
			return ct, nil
		}

		// Don't create master contact
		ct.ParentContact = masterContact.Id
		return ct, nil
	}

	return ct, nil
}
