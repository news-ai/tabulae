package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type UserInviteCode struct {
	Base

	InviteCode string `json:"invitecode"`
	Email      string `json:"email"`
	IsUsed     bool   `json:"isused"`

	ReferralUser int64 `json:"referraluser" apiModel:"User"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (uic *UserInviteCode) Create(c context.Context, r *http.Request) (*UserInviteCode, error) {
	// Create user
	uic.Created = time.Now()
	_, err := uic.Save(c)
	return uic, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (uic *UserInviteCode) Save(c context.Context) (*UserInviteCode, error) {
	uic.Updated = time.Now()
	uic.IsUsed = false

	k, err := nds.Put(c, uic.key(c, "UserInviteCode"), uic)
	if err != nil {
		return nil, err
	}
	uic.Id = k.IntID()
	return uic, nil
}

// Function to save a new user into App Engine
func (uic *UserInviteCode) Delete(c context.Context) (*UserInviteCode, error) {
	err := nds.Delete(c, uic.key(c, "UserInviteCode"))
	if err != nil {
		return nil, err
	}
	return uic, nil
}
