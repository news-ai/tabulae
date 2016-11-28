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

	// Social network settings
	LinkedinId      string `json:"-"`
	LinkedinAuthKey string `json:"-"`

	InstagramId      string `json:"instagramid"`
	InstagramAuthKey string `json:"-"`

	InvitedBy int64 `json:"-"`

	AgreeTermsAndConditions bool `json:"-"`
	EmailConfirmed          bool `json:"emailconfirmed"`

	GetDailyEmails bool `json:"getdailyemails"`

	BillingId int64 `json:"-"`
	TeamId    int64 `json:"teamid"`

	// Email settings
	EmailSignature string `json:"emailsignature"`

	Gmail           bool      `json:"gmail"`
	AccessToken     string    `json:"-"`
	GoogleCode      string    `json:"-"`
	GoogleExpiresIn time.Time `json:"-"`
	RefreshToken    string    `json:"-"`
	TokenType       string    `json:"-"`

	ExternalEmail bool  `json:"externalemail"`
	EmailSetting  int64 `json:"emailsetting"`

	SMTPValid    bool   `json:"-"`
	SMTPPassword []byte `json:"-"`

	IsAdmin  bool `json:"-"`
	IsActive bool `json:"isactive"`
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
	u.GetDailyEmails = true
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

func (u *User) SetStripeId(c context.Context, r *http.Request, currentUser User, stripeId string, stripePlanId string, isActive bool, isTrial bool) (*User, int64, error) {
	billing := Billing{}
	billing.StripeId = stripeId
	billing.StripePlanId = stripePlanId

	if isTrial {
		expires := time.Now().Add(time.Hour * 24 * 7 * time.Duration(1))
		billing.HasTrial = true
		billing.IsOnTrial = true
		billing.Expires = expires
	}

	_, err := billing.Create(c, r, currentUser)
	if err != nil {
		return u, 0, err
	}

	u.BillingId = billing.Id
	u.IsActive = isActive
	_, err = u.Save(c)
	if err != nil {
		return u, 0, err
	}
	return u, billing.Id, nil
}
