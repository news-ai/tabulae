package models

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type User struct {
	Base

	GoogleId string `json:"googleid"`

	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	Password []byte `json:"-"`
	ApiKey   string `json:"-"`

	Employers []int64 `json:"employers" apiModel:"Agency"`

	ResetPasswordCode string `json:"-"`
	ConfirmationCode  string `json:"-"`

	LastLoggedIn time.Time `json:"-"`

	LinkedinId      string `json:"-"`
	LinkedinAuthKey string `json:"-"`

	InstagramId      string `json:"-"`
	InstagramAuthKey string `json:"-"`

	StripeId       string `json:"-"`
	StripePlanId   string `json:"-"`
	IsBetaUser     bool   `json:"-"`
	StripeIsActive bool   `json:"isactive"`

	InvitedBy int64 `json:"-"`

	AgreeTermsAndConditions bool `json:"-"`
	EmailConfirmed          bool `json:"emailconfirmed"`
	IsAdmin                 bool `json:"-"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (u *User) Normalize() (*User, error) {
	u.Email = strings.ToLower(u.Email)
	u.FirstName = strings.Title(u.FirstName)
	u.LastName = strings.Title(u.LastName)
	return u, nil
}

func (u *User) Create(c context.Context, r *http.Request) (*User, error) {
	// Create user
	u.IsAdmin = false
	u.Created = time.Now()

	u.Normalize()

	_, err := u.Save(c)
	return u, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (u *User) Save(c context.Context) (*User, error) {
	u.Updated = time.Now()

	k, err := nds.Put(c, u.key(c, "User"), u)
	if err != nil {
		return nil, err
	}
	u.Id = k.IntID()
	return u, nil
}

func (u *User) ConfirmEmail(c context.Context) (*User, error) {
	u.EmailConfirmed = true
	u.ConfirmationCode = ""
	_, err := u.Save(c)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (u *User) ConfirmLoggedIn(c context.Context) (*User, error) {
	u.LastLoggedIn = time.Now()
	_, err := u.Save(c)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (u *User) SetStripeId(c context.Context, stripeId string, stripePlanId string, isBetaUser bool, isActive bool) (*User, error) {
	u.StripeId = stripeId
	u.StripePlanId = stripePlanId
	u.IsBetaUser = isBetaUser
	u.StripeIsActive = isActive
	_, err := u.Save(c)
	if err != nil {
		return u, err
	}
	return u, nil
}
