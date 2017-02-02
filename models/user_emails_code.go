package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type UserEmail struct {
	Email string `json:"email"`
}

type UserEmailCode struct {
	Base

	InviteCode string `json:"invitecode"`
	Email      string `json:"email"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (uec *UserEmailCode) Create(c context.Context, r *http.Request, currentUser User) (*UserEmailCode, error) {
	// Create user
	uec.CreatedBy = currentUser.Id
	uec.Created = time.Now()

	_, err := uec.Save(c)
	return uec, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (uec *UserEmailCode) Save(c context.Context) (*UserEmailCode, error) {
	uec.Updated = time.Now()
	k, err := nds.Put(c, uec.key(c, "UserEmailCode"), uec)
	if err != nil {
		return nil, err
	}
	uec.Id = k.IntID()
	return uec, nil
}

// Function to save a new user into App Engine
func (uec *UserEmailCode) Delete(c context.Context) (*UserEmailCode, error) {
	err := nds.Delete(c, uec.key(c, "UserEmailCode"))
	if err != nil {
		return nil, err
	}
	return uec, nil
}
