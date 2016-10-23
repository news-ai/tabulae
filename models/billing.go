package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Billing struct {
	Base

	StripeId     string    `json:"-"`
	StripePlanId string    `json:"-"`
	Expires      time.Time `json:"-"`
	HasTrial     bool      `json:"-"`
	IsOnTrial    bool      `json:"-"`

	CardsOnFile []string `json:"-"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (b *Billing) Create(c context.Context, r *http.Request, currentUser User) (*Billing, error) {
	b.CreatedBy = currentUser.Id
	b.Created = time.Now()
	_, err := b.Save(c)
	return b, err
}

/*
* Update methods
 */

// Function to save a new billing into App Engine
func (b *Billing) Save(c context.Context) (*Billing, error) {
	// Update the Updated time
	b.Updated = time.Now()

	k, err := nds.Put(c, b.key(c, "Billing"), b)
	if err != nil {
		return nil, err
	}
	b.Id = k.IntID()
	return b, nil
}
