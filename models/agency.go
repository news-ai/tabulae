package models

import (
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type Agency struct {
	Id int64 `json:"id" datastore:"-"`

	Name  string `json:"name"`
	Email string `json:"email"`

	Administrators []int64 `json:"administrators"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

/*
* Private methods
 */

// Generates a new key for the data to be stored on App Engine
func (a *Agency) key(c appengine.Context) *datastore.Key {
	if a.Id == 0 {
		a.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Agency", nil)
	}
	return datastore.NewKey(c, "Agency", "", a.Id, nil)
}

/*
* Get methods
 */

func getAgency(c appengine.Context, id int64) (Agency, error) {
	// Get the agency by id
	agencies := []Agency{}
	agencyId := datastore.NewKey(c, "Agency", "", id, nil)
	ks, err := datastore.NewQuery("Agency").Filter("__key__ =", agencyId).GetAll(c, &agencies)
	if err != nil {
		return Agency{}, err
	}
	if len(agencies) > 0 {
		agencies[0].Id = ks[0].IntID()
		return agencies[0], nil
	}
	return Agency{}, errors.New("No agency by this id")
}

/*
* Create methods
 */

func (a *Agency) create(c appengine.Context) (*Agency, error) {
	a.Created = time.Now()
	_, err := a.save(c)
	return a, err
}

/*
* Update methods
 */

// Function to save a new agency into App Engine
func (a *Agency) save(c appengine.Context) (*Agency, error) {
	k, err := datastore.Put(c, a.key(c), a)
	if err != nil {
		return nil, err
	}
	a.Id = k.IntID()
	return a, nil
}

/*
* Filter methods
 */

func filterAgency(c appengine.Context, queryType, query string) (Agency, error) {
	// Get an agency by their email extension
	agencies := []Agency{}
	ks, err := datastore.NewQuery("Publication").Filter(queryType+" =", query).GetAll(c, &agencies)
	if err != nil {
		return Agency{}, err
	}
	if len(agencies) > 0 {
		agencies[0].Id = ks[0].IntID()
		return agencies[0], nil
	}
	return Agency{}, errors.New("No agency by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

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

func GetAgency(c appengine.Context, id string) (Agency, error) {
	// Get the details of the current agency
	currentId, err := StringIdToInt(id)
	if err != nil {
		return Agency{}, err
	}

	agency, err := getAgency(c, currentId)
	if err != nil {
		return Agency{}, err
	}
	return agency, nil
}

/*
* Create methods
 */

func CreateAgencyFromUser(c appengine.Context, u *User) (Agency, error) {
	agencyEmail, err := GetAgencyEmail(u.Email)
	if err != nil {
		return Agency{}, err
	} else {
		agency, err := FilterAgencyByEmail(c, agencyEmail)
		if err != nil {
			agency = Agency{}
			agency.Name, err = GetAgencyName(agencyEmail)
			agency.Email = agencyEmail
			agency.Created = time.Now()
			agency.Administrators = append(agency.Administrators, u.Id)
			agency.create(c)
		}
		u.Employers = append(u.Employers, agency.Id)
		u.save(c)
		return agency, nil
	}
	return Agency{}, nil
}

/*
* Update methods
 */

func UpdateAgency(c appengine.Context, r *http.Request, id string) (Agency, error) {

}

/*
* Filter methods
 */

func FilterAgencyByEmail(c appengine.Context, email string) (Agency, error) {
	// Get the id of the current agency
	agency, err := filterAgency(c, "Email", email)
	if err != nil {
		return Agency{}, err
	}
	return agency, nil
}

/*
* Format methods
 */

func FormatAgencyId(c appengine.Context, agency *Agency) (int64, error) {
	// Get the id of the current agency
	agencyWithId, err := FilterAgencyByEmail(c, agency.Email)
	agency.Id = agencyWithId.Id

	if err != nil {
		return 0, err
	}

	return agency.Id, nil
}

func FormatAgenciesId(c appengine.Context, agencies []Agency) ([]int64, error) {
	// Get the id of the current agency
	agencyIds := []int64{}
	for i := 0; i < len(agencies); i++ {
		agencyId, err := FormatAgencyId(c, &agencies[i])

		if err != nil {
			return []int64{}, err
		}

		agencyIds = append(agencyIds, agencyId)
	}

	return agencyIds, nil
}
