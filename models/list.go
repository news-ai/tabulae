package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type CustomFieldsMap struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	CustomField bool   `json:"customfield"`
	Hidden      bool   `json:"hidden"`
}

type MediaList struct {
	Base

	Name   string `json:"name"`
	Client string `json:"client"`

	Contacts []int64 `json:"contacts"`

	FieldsMap []CustomFieldsMap `json:"fieldsmap"`

	FileUpload int64 `json:"fileupload"`

	Archived bool `json:"archived"`
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

// Function to save a new contact into App Engine
func (ml *MediaList) Save(c context.Context) (*MediaList, error) {
	// Update the Updated time
	ml.Updated = time.Now()

	k, err := nds.Put(c, ml.key(c, "MediaList"), ml)
	if err != nil {
		return nil, err
	}
	ml.Id = k.IntID()
	return ml, nil
}
