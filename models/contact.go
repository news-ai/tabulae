package models

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type Contact struct {
	Id int64 `json:"id" datastore:"-"`

	Name  string `json:"name"`
	Email string `json:"email"`

	WorksAt []Publication `json:"worksat" datastore:"-"`

	CreatedBy User `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
}

/*
* Private methods
 */

// Code to get data from App Engine
func defaultContactList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "ContactList", "default", 0, nil)
}

// Generates a new key for the data to be stored on App Engine
func (ct *Contact) key(c appengine.Context) *datastore.Key {
	if ct.Id == 0 {
		ct.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Contact", defaultContactList(c))
	}
	return datastore.NewKey(c, "Contact", "", ct.Id, defaultContactList(c))
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

// func CreateContactFromWorksAt(c appengine.Context, u *User) (Agency, error) {
// }
