package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Notification struct {
	Base
}

type NotificationObject struct {
	Base

	NoticationId int64

	Object   string
	ObjectId int64
}

type NotificationChange struct {
	Base

	NoticationObjectId int64

	Verb  string
	Actor string

	Message string `json:"message"`

	Read bool
}

/*
* Notification Public methods
 */

/*
* Create methods
 */

func (n *Notification) Create(c context.Context, currentUser User) (*Notification, error) {
	n.CreatedBy = currentUser.Id
	n.Created = time.Now()

	_, err := n.Save(c)
	return n, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (n *Notification) Save(c context.Context) (*Notification, error) {
	// Update the Updated time
	n.Updated = time.Now()

	k, err := nds.Put(c, n.key(c, "Notification"), n)
	if err != nil {
		return nil, err
	}
	n.Id = k.IntID()
	return n, nil
}

/*
* NotificationObject Public methods
 */

/*
* Create methods
 */

func (no *NotificationObject) Create(c context.Context, currentUser User) (*NotificationObject, error) {
	no.CreatedBy = currentUser.Id
	no.Created = time.Now()

	_, err := no.Save(c)
	return no, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (no *NotificationObject) Save(c context.Context) (*NotificationObject, error) {
	// Update the Updated time
	no.Updated = time.Now()

	k, err := nds.Put(c, no.key(c, "NotificationObject"), no)
	if err != nil {
		return nil, err
	}
	no.Id = k.IntID()
	return no, nil
}

/*
* NotificationChange Public methods
 */

/*
* Create methods
 */

func (nc *NotificationChange) Create(c context.Context, r *http.Request, currentUser User) (*NotificationChange, error) {
	nc.CreatedBy = currentUser.Id
	nc.Created = time.Now()
	nc.Read = false

	_, err := nc.Save(c)
	return nc, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (nc *NotificationChange) Save(c context.Context) (*NotificationChange, error) {
	// Update the Updated time
	nc.Updated = time.Now()

	k, err := nds.Put(c, nc.key(c, "NotificationChange"), nc)
	if err != nil {
		return nil, err
	}
	nc.Id = k.IntID()
	return nc, nil
}
