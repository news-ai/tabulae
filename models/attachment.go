package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Attachment struct {
	Base

	FileName string `json:"filename"`
	EmailId  int64  `json:"emailid"`

	FileExists bool `json:"fileexists"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (at *Attachment) Create(c context.Context, r *http.Request, currentUser User) (*Attachment, error) {
	at.CreatedBy = currentUser.Id
	at.Created = time.Now()
	_, err := at.Save(c)
	return at, err
}

/*
* Update methods
 */

// Function to save a new agency into App Engine
func (at *Attachment) Save(c context.Context) (*Attachment, error) {
	// Update the Updated time
	at.Updated = time.Now()

	k, err := nds.Put(c, at.key(c, "Attachment"), at)
	if err != nil {
		return nil, err
	}
	at.Id = k.IntID()
	return at, nil
}
