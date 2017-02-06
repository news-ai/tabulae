package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type BillingHistory struct {
	Base
}

/*
* Public methods
 */

/*
* Create methods
 */

func (bh *BillingHistory) Create(c context.Context, r *http.Request, currentUser User) (*BillingHistory, error) {
	bh.CreatedBy = currentUser.Id
	bh.Created = time.Now()
	_, err := bh.Save(c)
	return bh, err
}

/*
* Update methods
 */

// Function to save a new billing into App Engine
func (bh *BillingHistory) Save(c context.Context) (*BillingHistory, error) {
	// Update the Updated time
	bh.Updated = time.Now()

	k, err := nds.Put(c, bh.key(c, "BillingHistory"), bh)
	if err != nil {
		return nil, err
	}
	bh.Id = k.IntID()
	return bh, nil
}
