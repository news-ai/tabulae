package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Agency struct {
	Base

	Name  string `json:"name"`
	Email string `json:"email"`

	Administrators []int64 `json:"administrators"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (a *Agency) Create(c context.Context, r *http.Request, currentUser User) (*Agency, error) {
	a.CreatedBy = currentUser.Id
	a.Created = time.Now()
	_, err := a.Save(c)
	return a, err
}

/*
* Update methods
 */

// Function to save a new agency into App Engine
func (a *Agency) Save(c context.Context) (*Agency, error) {
	// Update the Updated time
	a.Updated = time.Now()

	k, err := nds.Put(c, a.key(c, "Agency"), a)
	if err != nil {
		return nil, err
	}
	a.Id = k.IntID()
	return a, nil
}
