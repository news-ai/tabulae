package models

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type CustomContactField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Contact struct {
	Id int64 `json:"id" datastore:"-"`

	// Contact information
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`

	// Notes on a particular contact
	Notes string `json:"notes"`

	// Publications this contact works for now and before
	Employers     []int64 `json:"employers"`
	PastEmployers []int64 `json:"pastemployers"`

	// Social information
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	MuckRack  string `json:"-"`
	Website   string `json:"website"`
	Blog      string `json:"blog"`

	// Custom fields
	CustomFields []CustomContactField `json:"customfields"`

	// Parent contact
	IsMasterContact bool  `json:"ismastercontact"`
	ParentContact   int64 `json:"parent"`

	// Is information outdated
	IsOutdated bool `json:"isoutdated"`

	CreatedBy int64 `json:"createdby"`

	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
	LinkedInUpdated time.Time `json:"linkedinupdated"`
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
* Create methods
 */

func (ct *Contact) Create(c appengine.Context, r *http.Request, currentUser User) (*Contact, error) {
	ct.CreatedBy = currentUser.Id
	ct.Created = time.Now()
	ct.Normalize()

	_, err := ct.Save(c, r)
	return ct, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ct *Contact) Save(c appengine.Context, r *http.Request) (*Contact, error) {
	// Update the Updated time
	ct.Updated = time.Now()

	k, err := datastore.Put(c, ct.key(c), ct)
	if err != nil {
		return nil, err
	}
	ct.Id = k.IntID()
	return ct, nil
}

/*
* Normalization methods
 */

func (ct *Contact) Normalize() (*Contact, error) {
	ct.LinkedIn = stripQueryString(ct.LinkedIn)
	ct.Twitter = stripQueryString(ct.Twitter)
	ct.Instagram = stripQueryString(ct.Instagram)
	ct.MuckRack = stripQueryString(ct.MuckRack)
	ct.Website = stripQueryString(ct.Website)
	ct.Blog = stripQueryString(ct.Blog)

	return ct, nil
}
