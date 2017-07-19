package models

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/news-ai/web/utilities"

	apiModels "github.com/news-ai/api/models"

	"github.com/qedus/nds"
)

type ContactV2 struct {
	apiModels.Base

	TeamId int64 `json:"teamid"`

	// Contact information
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`

	// Which lists is the contact on
	ListIds []int64 `json:"listids" apiModel:"MediaList"`

	// Notes on a particular contact
	Notes string `json:"notes"`

	// Publications this contact works for now and before
	Employers     []int64 `json:"employers" apiModel:"Publication"`
	PastEmployers []int64 `json:"pastemployers" apiModel:"Publication"`

	ImageURLs []string `json:"imageurls" datastore:",noindex"`

	Tags []string `json:"tags"`

	// Social + personal information
	LinkedIn    string `json:"linkedin"`
	Twitter     string `json:"twitter"`
	Instagram   string `json:"instagram"`
	Website     string `json:"website"`
	Blog        string `json:"blog"`
	Location    string `json:"location"`
	PhoneNumber string `json:"phonenumber"`

	// Custom fields
	CustomFields []CustomContactField `json:"customfields" datastore:",noindex"`

	/*
		Fields not stored in datastore
	*/

	// Is information outdated
	IsOutdated   bool `json:"isoutdated" datastore:"-"`
	EmailBounced bool `json:"emailbounced" datastore:"-"`

	ReadOnly bool `json:"readonly" datastore:"-"`
}

/*
* Public methods
 */

func (ct *ContactV2) Key(c context.Context) *datastore.Key {
	return ct.BaseKey(c, "ContactV2")
}

/*
* Create methods
 */

func (ct *ContactV2) Create(c context.Context, r *http.Request, currentUser apiModels.User) (*ContactV2, error) {
	ct.CreatedBy = currentUser.Id
	ct.Created = time.Now()
	ct.Normalize()
	ct.FormatName()

	_, err := ct.Save(c, r)
	return ct, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ct *ContactV2) Save(c context.Context, r *http.Request) (*ContactV2, error) {
	// Update the Updated time
	ct.Updated = time.Now()
	ct.Normalize()
	ct.FormatName()

	k, err := nds.Put(c, ct.BaseKey(c, "ContactV2"), ct)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	ct.Id = k.IntID()
	return ct, nil
}

/*
* Normalization methods
 */

func (ct *ContactV2) FormatName() (*ContactV2, error) {
	if len(ct.FirstName) > 0 && len(ct.LastName) == 0 {
		if strings.Contains(ct.FirstName, ",") {
			nameSplit := strings.Split(ct.FirstName, ", ")
			if len(nameSplit) > 1 {
				// Agarwal, Abhi as First Name
				ct.FirstName = nameSplit[1]
				ct.LastName = nameSplit[0]
			}
		} else if strings.Contains(ct.FirstName, " ") {
			// Abhi Agarwal as First Name
			nameSplit := strings.Split(ct.FirstName, " ")
			if len(nameSplit) > 1 {
				ct.FirstName = nameSplit[0]
				ct.LastName = strings.Join(nameSplit[1:], " ")
			}
		}
	}

	return ct, nil
}

func (ct *ContactV2) Normalize() (*ContactV2, error) {
	ct.LinkedIn = strings.ToLower(utilities.StripQueryString(ct.LinkedIn))
	ct.Twitter = utilities.StripQueryString(ct.Twitter)
	ct.Twitter = strings.ToLower(utilities.NormalizeUrlToUsername(ct.Twitter, "twitter.com"))
	ct.Instagram = utilities.StripQueryString(ct.Instagram)
	ct.Instagram = strings.ToLower(utilities.NormalizeUrlToUsername(ct.Instagram, "instagram.com"))

	ct.Website = utilities.StripQueryStringForWebsite(ct.Website)
	ct.Blog = utilities.StripQueryStringForWebsite(ct.Blog)

	ct.Email = strings.ToLower(ct.Email)
	ct.Email = strings.TrimSpace(ct.Email)

	return ct, nil
}

/*
* Action methods
 */

func (ct *ContactV2) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(ct, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
