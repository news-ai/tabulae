package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Email struct {
	Base

	// Which list it belongs to
	ListId     int64 `json:"listid" apiModel:"List"`
	TemplateId int64 `json:"templateid" apiModel:"Template"`
	ContactId  int64 `json:"contactId" apiModel:"Contact"`

	Sender  string `json:"sender"`
	To      string `json:"to"`
	Subject string `json:"subject" datastore:",noindex"`
	Body    string `json:"body" datastore:",noindex"`

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	SendAt time.Time `json:"sendat"`

	SendGridId string `json:"-"`

	Attachments []int64 `json:"attachments" datastore:",noindex" apiModel:"File"`

	Delievered    bool   `json:"delivered"`
	BouncedReason string `json:"bouncedreason"`
	Bounced       bool   `json:"bounced"`
	Clicked       int    `json:"clicked"`
	Opened        int    `json:"opened"`
	Spam          bool   `json:"spam"`
	Cancel        bool   `json:"cancel"`

	IsSent bool `json:"issent"`
}

/*
* Private methods
 */

/*
* Create methods
 */

func (e *Email) Create(c context.Context, r *http.Request, currentUser User) (*Email, error) {
	e.IsSent = false
	e.CreatedBy = currentUser.Id
	e.Created = time.Now()

	_, err := e.Save(c)
	return e, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (e *Email) Save(c context.Context) (*Email, error) {
	// Update the Updated time
	e.Updated = time.Now()

	k, err := nds.Put(c, e.key(c, "Email"), e)
	if err != nil {
		return nil, err
	}
	e.Id = k.IntID()
	return e, nil
}

func (e *Email) MarkSent(c context.Context, emailId string) (*Email, error) {
	e.IsSent = true
	e.SendGridId = emailId
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (e *Email) MarkBounced(c context.Context, reason string) (*Email, error) {
	e.Bounced = true
	e.BouncedReason = reason
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (e *Email) MarkClicked(c context.Context) (*Email, error) {
	e.Clicked += 1
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (e *Email) MarkDelivered(c context.Context) (*Email, error) {
	e.Delievered = true
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (e *Email) MarkSpam(c context.Context) (*Email, error) {
	e.Spam = true
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (e *Email) MarkOpened(c context.Context) (*Email, error) {
	e.Opened += 1
	_, err := e.Save(c)
	if err != nil {
		return e, err
	}
	return e, nil
}
