package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Contact struct {
	Id int64 `json:"id" datastore:"-"`

	Name  string `json:"name"`
	Email string `json:"email"`

	WorksAt []Publication `json:"worksat"`

	CreatedBy User `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (ct *Contact) key(c appengine.Context) *datastore.Key {
	if ct.Id == 0 {
		ct.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Contact", nil)
	}
	return datastore.NewKey(c, "Contact", "", ct.Id, nil)
}

/*
* Get methods
 */

func getContact(c appengine.Context, id string) (Contact, error) {
	// Get the Contact by id
	contacts := []Contact{}
	ks, err := datastore.NewQuery("Contact").Filter("ID =", id).GetAll(c, &contacts)
	if err != nil {
		return Contact{}, err
	}
	if len(contacts) > 0 {
		contacts[0].Id = ks[0].IntID()
		return contacts[0], nil
	}
	return Contact{}, errors.New("No contact by this id")
}

// func getContactByWorksAt(c appengine.Context, worksat string) (Contact, error) {

// }

/*
* Create methods
 */

func (ct *Contact) create(c appengine.Context) (*Contact, error) {
	currentUser, err := GetCurrentUser(c)
	if err != nil {
		return ct, err
	}

	ct.CreatedBy = currentUser
	ct.Created = time.Now()

	_, err = ct.save(c)
	return ct, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ct *Contact) save(c appengine.Context) (*Contact, error) {
	k, err := datastore.Put(c, ct.key(c), ct)
	if err != nil {
		return nil, err
	}
	ct.Id = k.IntID()
	return ct, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single contact
func GetContacts(c appengine.Context) ([]Contact, error) {
	contacts := []Contact{}
	ks, err := datastore.NewQuery("Contact").GetAll(c, &contacts)
	if err != nil {
		return []Contact{}, err
	}
	for i := 0; i < len(contacts); i++ {
		contacts[i].Id = ks[i].IntID()
	}
	return contacts, nil
}

// func GetContactByWorksAt(c appengine.Context, email string) (Contact, error) {
// }

func GetContact(c appengine.Context, id string) (Contact, error) {
	// Get the details of the current user
	contact, err := getContact(c, id)
	if err != nil {
		return Contact{}, err
	}
	return contact, nil
}

/*
* Create methods
 */

// Method not completed
func CreateContact(c appengine.Context, w http.ResponseWriter, r *http.Request) (Contact, error) {
	decoder := json.NewDecoder(r.Body)
	var contact Contact
	err := decoder.Decode(&contact)
	if err != nil {
		return Contact{}, err
	}

	// WorksAt
	for i := 0; i < len(contact.WorksAt); i++ {
		publication, err := GetPublication(c, contact.WorksAt[i].Id)
		if err != nil {
			return Contact{}, err
		}
		contact.WorksAt[i] = publication
	}

	// Create contact
	_, err = contact.create(c)
	if err != nil {
		return Contact{}, err
	}

	return contact, nil
}

// func CreateContactFromWorksAt(c appengine.Context, u *User) (Agency, error) {
// }
