package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Team struct {
	Base

	Members []int64 `json:"members" apiModel:"User"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new team into App Engine
func (t *Team) Create(c context.Context, r *http.Request, currentUser User) (*Team, error) {
	t.Members = append(t.Members, currentUser.Id)
	t.CreatedBy = currentUser.Id
	t.Created = time.Now()

	_, err := t.Save(c)
	return t, err
}

/*
* Update methods
 */

// Function to save a new team into App Engine
func (t *Team) Save(c context.Context) (*Team, error) {
	// Update the Updated time
	t.Updated = time.Now()

	// Save the object
	k, err := nds.Put(c, t.key(c, "Team"), t)
	if err != nil {
		return nil, err
	}
	t.Id = k.IntID()
	return t, nil
}
