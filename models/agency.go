package models

import (
	"errors"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type Agency struct {
	Id int64 `json:"id" datastore:"-"`

	Name  string `json:"name"`
	Email string `json:"email"`

	CreatedBy int64 `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
}

// Code to get data from App Engine
func defaultAgencyList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "AgencyList", "default", 0, nil)
}

// Generates a new key for the data to be stored on App Engine
func (a *Agency) key(c appengine.Context) *datastore.Key {
	if a.Id == 0 {
		a.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Agency", defaultAgencyList(c))
	}
	return datastore.NewKey(c, "Agency", "", a.Id, defaultAgencyList(c))
}

// Function to save a new agency into App Engine
func (a *Agency) save(c appengine.Context) (*Agency, error) {
	k, err := datastore.Put(c, a.key(c), a)
	if err != nil {
		return nil, err
	}
	a.Id = k.IntID()
	return a, nil
}

func (a *Agency) create(c appengine.Context) (*Agency, error) {
	currentUser := user.Current(c)
	currentUserId, err := strconv.ParseInt(currentUser.ID, 10, 32)
	if err != nil {
		return &Agency{}, err
	}
	a.CreatedBy = currentUserId
	_, err = a.save(c)
	return a, err
}

// Gets every single agency
func GetAgencies(c appengine.Context) ([]Agency, error) {
	agencies := []Agency{}
	ks, err := datastore.NewQuery("Agency").GetAll(c, &agencies)
	if err != nil {
		return []Agency{}, err
	}
	for i := 0; i < len(agencies); i++ {
		agencies[i].Id = ks[i].IntID()
	}
	return agencies, nil
}

func getAgency(c appengine.Context, id string) (Agency, error) {
	// Get the current signed in user details by Email
	agencies := []Agency{}
	ks, err := datastore.NewQuery("Agency").Filter("ID =", id).GetAll(c, &agencies)
	if err != nil {
		return Agency{}, err
	}
	if len(agencies) > 0 {
		agencies[0].Id = ks[0].IntID()
		return agencies[0], nil
	}
	return Agency{}, errors.New("No agency by this id")
}

func GetAgency(c appengine.Context, id string) (Agency, error) {
	// Get the details of the current user
	agency, err := getAgency(c, id)
	if err != nil {
		return Agency{}, err
	}
	return agency, nil
}
