package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	"github.com/qedus/nds"
)

type ContactUnsubscribe struct {
	apiModels.Base

	Email        string `json:"email"`
	Unsubscribed bool   `json:"unsubscribed"`
}

/*
* Public methods
 */

func (cu *ContactUnsubscribe) Key(c context.Context) *datastore.Key {
	return cu.key(c, "ContactUnsubscribe")
}

/*
* Create methods
 */

func (cu *ContactUnsubscribe) Create(c context.Context, r *http.Request, currentUser User) (*ContactUnsubscribe, error) {
	cu.CreatedBy = currentUser.Id
	cu.Created = time.Now()

	_, err := cu.Save(c, r)
	return cu, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (cu *ContactUnsubscribe) Save(c context.Context, r *http.Request) (*ContactUnsubscribe, error) {
	// Update the Updated time
	cu.Updated = time.Now()

	k, err := nds.Put(c, cu.key(c, "ContactUnsubscribe"), cu)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	cu.Id = k.IntID()
	return cu, nil
}
