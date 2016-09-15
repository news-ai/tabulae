package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type UserToken struct {
	Base

	Token        string `json:"token"`
	ChannelToken string `json:"channeltoken"`
	IsUsed       bool   `json:"isused"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (ut *UserToken) Create(c context.Context, r *http.Request) (*UserToken, error) {
	// Create user
	ut.Created = time.Now()
	_, err := ut.Save(c)
	return ut, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (ut *UserToken) Save(c context.Context) (*UserToken, error) {
	ut.Updated = time.Now()

	k, err := nds.Put(c, ut.key(c, "UserToken"), ut)
	if err != nil {
		return nil, err
	}
	ut.Id = k.IntID()
	return ut, nil
}

// Function to save a new user into App Engine
func (ut *UserToken) Delete(c context.Context) (*UserToken, error) {
	err := nds.Delete(c, ut.key(c, "UserToken"))
	if err != nil {
		return nil, err
	}
	return ut, nil
}