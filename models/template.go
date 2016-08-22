package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Template struct {
	Base

	Subject string `json:"subject" datastore:",noindex"`
	Body    string `json:"body" datastore:",noindex"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new team into App Engine
func (tpl *Template) Create(c context.Context, r *http.Request, currentUser User) (*Template, error) {
	tpl.CreatedBy = currentUser.Id
	tpl.Created = time.Now()

	_, err := tpl.Save(c)
	return tpl, err
}

/*
* Update methods
 */

// Function to save a new team into App Engine
func (tpl *Template) Save(c context.Context) (*Template, error) {
	// Update the Updated time
	tpl.Updated = time.Now()

	// Save the object
	k, err := nds.Put(c, tpl.key(c, "Template"), tpl)
	if err != nil {
		return nil, err
	}
	tpl.Id = k.IntID()
	return tpl, nil
}
