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
	ListId int64 `json:"listid"`

	Sender  string `json:"sender"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	SendGridId string `json:"-"`

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
