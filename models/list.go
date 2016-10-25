package models

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/qedus/nds"
)

type CustomFieldsMap struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	CustomField bool   `json:"customfield"`
	Hidden      bool   `json:"hidden"`
	Internal    bool   `json:"internal" datastore:"-"`
	ReadOnly    bool   `json:"readonly" datastore:"-"`
}

type MediaList struct {
	Base

	Name   string `json:"name"`
	Client string `json:"client"`

	Contacts []int64 `json:"contacts" apiModel:"Contact"`

	FieldsMap []CustomFieldsMap `json:"fieldsmap" datastore:",noindex"`

	CustomFields []string `json:"-" datastore:",noindex"`
	Fields       []string `json:"-" datastore:",noindex"`

	FileUpload int64 `json:"fileupload" apiModel:"File"`

	ReadOnly   bool `json:"readonly" datastore:"-"`
	PublicList bool `json:"publiclist"`
	Archived   bool `json:"archived"`
	Subscribed bool `json:"subscribed"`
}

/*
* Private methods
 */

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
	}

	for i := 0; i < len(ml.FieldsMap); i++ {
		if strings.Contains(ml.FieldsMap[i].Name, "instagram") {
			if _, ok := newFieldsMap[ml.FieldsMap[i].Name]; ok {
				newFieldsMap[ml.FieldsMap[i].Name] = false
			}
		}
	}

	isChanged := false
	for key, v := range newFieldsMap {
		if v {
			isChanged = true
			field := CustomFieldsMap{
				Name:        key,
				Value:       key,
				CustomField: true,
				Hidden:      true,
			}
			ml.FieldsMap = append(ml.FieldsMap, field)
		}
	}

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
		return nil, err
	}
	ml.Format(k, "lists")
	return ml, nil
}

func (ml *MediaList) Format(key *datastore.Key, modelType string) {
	ml.Type = modelType
	ml.Id = key.IntID()

	for i := 0; i < len(ml.FieldsMap); i++ {
		if ml.FieldsMap[i].Name == "employers" || ml.FieldsMap[i].Name == "pastemployers" {
			ml.FieldsMap[i].Internal = true
		}

		if ml.FieldsMap[i].Name == "instagramfollowers" || ml.FieldsMap[i].Name == "instagramfollowing" || ml.FieldsMap[i].Name == "instagramlikes" || ml.FieldsMap[i].Name == "instagramcomments" || ml.FieldsMap[i].Name == "instagramposts" {
			ml.FieldsMap[i].ReadOnly = true
		}
	}
}
