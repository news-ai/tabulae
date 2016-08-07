package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type Publication struct {
	Base

	Name string `json:"name"`
	Url  string `json:"url"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new publication into App Engine
func (p *Publication) Create(c context.Context, r *http.Request, currentUser User) (*Publication, error) {
	p.CreatedBy = currentUser.Id
	p.Created = time.Now()

	_, err := p.Save(c)
	return p, err
}

/*
* Update methods
 */

// Function to save a new publication into App Engine
func (p *Publication) Save(c context.Context) (*Publication, error) {
	// Update the Updated time
	p.Updated = time.Now()

	// Save the object
	k, err := datastore.Put(c, p.key(c, "Publication"), p)
	if err != nil {
		return nil, err
	}
	p.Id = k.IntID()
	return p, nil
}
