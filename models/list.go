package models

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"
)

type CustomFieldsMap struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	CustomField bool   `json:"customfield"`
	Hidden      bool   `json:"hidden"`
	Internal    bool   `json:"internal" datastore:"-"`
	ReadOnly    bool   `json:"readonly" datastore:"-"`
	Description string `json:"description" datastore:"-"`
	Type        string `json:"type" datastore:"-"`
}

type MediaList struct {
	Base

	Name   string `json:"name"`
	Client string `json:"client"`

	Contacts []int64 `json:"contacts" apiModel:"Contact"`

	FieldsMap []CustomFieldsMap `json:"fieldsmap" datastore:",noindex"`

	Tags []string `json:"tags" datastore:",noindex"`

	CustomFields []string `json:"-" datastore:",noindex"`
	Fields       []string `json:"-" datastore:",noindex"`

	FileUpload int64 `json:"fileupload" apiModel:"File"`

	TeamId int64 `json:"teamid"`

	ReadOnly   bool `json:"readonly" datastore:"-"`
	PublicList bool `json:"publiclist"`
	Archived   bool `json:"archived"`
	Subscribed bool `json:"subscribed"`

	IsDeleted bool `json:"isdeleted"`
}

/*
* Private variables
 */

var fieldsMapValueToDescription = map[string]string{
	"instagramfollowers": "Updated on a daily basis",
	"instagramfollowing": "Updated on a daily basis",
	"instagramlikes":     "Updated on a daily basis",
	"instagramcomments":  "Updated on a daily basis",
	"instagramposts":     "Updated on a daily basis",

	"twitterfollowers": "Updated on a daily basis",
	"twitterfollowing": "Updated on a daily basis",
	"twitterlikes":     "Updated on a daily basis",
	"twitterretweets":  "Updated on a daily basis",
	"twitterposts":     "Updated on a daily basis",

	"latestheadline": "Updated on a daily basis",
}

/*
* Public methods
 */

/*
* Create methods
 */

func (ml *MediaList) Create(c context.Context, r *http.Request, currentUser User) (*MediaList, error) {
	ml.CreatedBy = currentUser.Id
	ml.Created = time.Now()

	_, err := ml.Save(c)
	return ml, err
}

/*
* Update methods
 */

func (ml *MediaList) AddNewCustomFieldsMapToOldLists(c context.Context) {
	newFieldsMap := map[string]bool{
		"instagramfollowers": true,
		"instagramfollowing": true,
		"instagramlikes":     true,
		"instagramcomments":  true,
		"instagramposts":     true,

		"twitterfollowers": true,
		"twitterfollowing": true,
		"twitterlikes":     true,
		"twitterretweets":  true,
		"twitterposts":     true,

		"latestheadline": true,
		"lastcontacted":  true,
	}

	newDefaultFieldsMap := map[string]bool{
		"phonenumber": true,
	}

	newFieldsMapNames := map[string]string{
		"instagramfollowers": "Instagram Followers",
		"instagramfollowing": "Instagram Following",
		"instagramlikes":     "Instagram Likes",
		"instagramcomments":  "Instagram Comments",
		"instagramposts":     "Instagram Posts",

		"twitterfollowers": "Twitter Followers",
		"twitterfollowing": "Twitter Following",
		"twitterlikes":     "Twitter Likes",
		"twitterretweets":  "Twitter Retweets",
		"twitterposts":     "Twitter Posts",

		"latestheadline": "Latest Headline",

		"lastcontacted": "Last Contacted",

		"firstname":     "First Name",
		"lastname":      "Last Name",
		"email":         "Email",
		"employers":     "Employers",
		"pastemployers": "Past Employers",
		"notes":         "Notes",
		"linkedin":      "Linkedin",
		"twitter":       "Twitter",
		"instagram":     "Instagram",
		"website":       "Website",
		"blog":          "Blog",
		"phonenumber":   "Phone #",
		"tags":          "Tags",
	}

	isChanged := false

	for i := 0; i < len(ml.FieldsMap); i++ {
		if strings.Contains(ml.FieldsMap[i].Value, "instagram") {
			if _, ok := newFieldsMap[ml.FieldsMap[i].Value]; ok {
				newFieldsMap[ml.FieldsMap[i].Value] = false
			}
		}
		if strings.Contains(ml.FieldsMap[i].Value, "twitter") {
			if _, ok := newFieldsMap[ml.FieldsMap[i].Value]; ok {
				newFieldsMap[ml.FieldsMap[i].Value] = false
			}
		}

		if strings.Contains(ml.FieldsMap[i].Value, "latestheadline") {
			if _, ok := newFieldsMap[ml.FieldsMap[i].Value]; ok {
				newFieldsMap[ml.FieldsMap[i].Value] = false
			}
		}

		if strings.Contains(ml.FieldsMap[i].Value, "lastcontacted") {
			if _, ok := newFieldsMap[ml.FieldsMap[i].Value]; ok {
				newFieldsMap[ml.FieldsMap[i].Value] = false
			}
		}

		if _, ok := newDefaultFieldsMap[ml.FieldsMap[i].Value]; ok {
			newDefaultFieldsMap[ml.FieldsMap[i].Value] = false
		}

		// If this particular name exists in newFieldsMapNames
		if _, ok := newFieldsMapNames[ml.FieldsMap[i].Name]; ok {
			ml.FieldsMap[i].Name = newFieldsMapNames[ml.FieldsMap[i].Name]
			isChanged = true
		}
	}

	for key, v := range newFieldsMap {
		if v {
			isChanged = true
			field := CustomFieldsMap{
				Name:        newFieldsMapNames[key],
				Value:       key,
				CustomField: true,
				Hidden:      true,
			}
			ml.FieldsMap = append(ml.FieldsMap, field)
		}
	}

	for key, v := range newDefaultFieldsMap {
		if v {
			isChanged = true
			field := CustomFieldsMap{
				Name:        newFieldsMapNames[key],
				Value:       key,
				CustomField: false,
				Hidden:      true,
			}
			ml.FieldsMap = append(ml.FieldsMap, field)
		}
	}

	// Remove Duplicates
	duplicateValues := map[string]bool{}
	finalFieldsMap := []CustomFieldsMap{}
	for i := 0; i < len(ml.FieldsMap); i++ {
		if _, ok := duplicateValues[ml.FieldsMap[i].Value]; !ok {
			finalFieldsMap = append(finalFieldsMap, ml.FieldsMap[i])
			duplicateValues[ml.FieldsMap[i].Value] = true
		}
	}
	ml.FieldsMap = finalFieldsMap

	if isChanged {
		ml.Save(c)
	}
}

// Function to save a new contact into App Engine
func (ml *MediaList) Save(c context.Context) (*MediaList, error) {
	// Update the Updated time
	ml.Updated = time.Now()

	k, err := nds.Put(c, ml.key(c, "MediaList"), ml)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	ml.Format(k, "lists")
	return ml, nil
}

func (ml *MediaList) Format(key *datastore.Key, modelType string) {
	ml.Type = modelType
	ml.Id = key.IntID()

	// Add descriptions on runtime
	for i := 0; i < len(ml.FieldsMap); i++ {
		if ml.FieldsMap[i].Value == "employers" || ml.FieldsMap[i].Value == "pastemployers" {
			ml.FieldsMap[i].Internal = true
		}

		if ml.FieldsMap[i].Value != "twitter" && strings.Contains(ml.FieldsMap[i].Value, "twitter") {
			ml.FieldsMap[i].ReadOnly = true
		}

		if ml.FieldsMap[i].Value != "instagram" && strings.Contains(ml.FieldsMap[i].Value, "instagram") {
			ml.FieldsMap[i].ReadOnly = true
		}

		if ml.FieldsMap[i].Value == "latestheadline" || ml.FieldsMap[i].Value == "lastcontacted" {
			ml.FieldsMap[i].ReadOnly = true
		}

		if ml.FieldsMap[i].Value == "lastcontacted" {
			ml.FieldsMap[i].Type = "Date"
		}

		// If this particular value exists in fieldsMapValueToDescription then add description
		if val, ok := fieldsMapValueToDescription[ml.FieldsMap[i].Value]; ok {
			ml.FieldsMap[i].Description = val
		}
	}
}
