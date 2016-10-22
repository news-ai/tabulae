package models

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/news-ai/web/utilities"

	"github.com/qedus/nds"
)

type CustomContactField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Contact struct {
	Base

	// Contact information
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`

	ListId int64 `json:"listid" apiModel:"MediaList"`

	// Notes on a particular contact
	Notes string `json:"notes"`

	// Publications this contact works for now and before
	Employers     []int64 `json:"employers" apiModel:"Publication"`
	PastEmployers []int64 `json:"pastemployers" apiModel:"Publication"`

	// Social information
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	MuckRack  string `json:"-"`
	Website   string `json:"website"`
	Blog      string `json:"blog"`

	TwitterInvalid   bool `json:"twitterinvalid"`
	InstagramInvalid bool `json:"instagraminvalid"`

	TwitterPrivate   bool `json:"twitterprivate"`
	InstagramPrivate bool `json:"instagramprivate"`

	Location    string `json:"location"`
	PhoneNumber string `json:"phonenumber"`

	// Custom fields
	CustomFields []CustomContactField `json:"customfields" datastore:",noindex"`

	// Parent contact
	IsMasterContact bool  `json:"ismastercontact"`
	ParentContact   int64 `json:"parent" apiModel:"Contact"`

	// Is information outdated
	IsOutdated   bool `json:"isoutdated"`
	EmailBounced bool `json:"emailbounced"`

	IsDeleted bool `json:"isdeleted"`

	LinkedInUpdated time.Time `json:"linkedinupdated"`
}

/*
* Public methods
 */

func (ct *Contact) Key(c context.Context) *datastore.Key {
	return ct.key(c, "Contact")
}

/*
* Create methods
 */

func (ct *Contact) Create(c context.Context, r *http.Request, currentUser User) (*Contact, error) {
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
func (ct *Contact) Save(c context.Context, r *http.Request) (*Contact, error) {
	// Update the Updated time
	ct.Updated = time.Now()
	ct.Normalize()

	k, err := nds.Put(c, ct.key(c, "Contact"), ct)
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
	ct.LinkedIn = strings.ToLower(utilities.StripQueryString(ct.LinkedIn))
	ct.Twitter = utilities.StripQueryString(ct.Twitter)
	ct.Twitter = strings.ToLower(utilities.NormalizeUrlToUsername(ct.Twitter, "twitter.com"))
	ct.Instagram = utilities.StripQueryString(ct.Instagram)
	ct.Instagram = strings.ToLower(utilities.NormalizeUrlToUsername(ct.Instagram, "instagram.com"))
	ct.MuckRack = strings.ToLower(utilities.StripQueryString(ct.MuckRack))

	ct.Website = strings.ToLower(utilities.StripQueryStringForWebsite(ct.Website))
	ct.Blog = strings.ToLower(utilities.StripQueryStringForWebsite(ct.Blog))

	return ct, nil
}

/*
* Action methods
 */

func (c *Contact) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := SetField(c, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
