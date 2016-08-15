package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type File struct {
	Base

	FileName string `json:"filename"`
	ListId   int64  `json:"listid"`

	Order []string `json:"order"`

	Imported   bool `json:"imported"`
	FileExists bool `json:"fileexists"`
}

type FileOrder struct {
	Order []string `json:"order"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (f *File) Create(c context.Context, r *http.Request, currentUser User) (*File, error) {
	f.CreatedBy = currentUser.Id
	f.Created = time.Now()

	_, err := f.Save(c)
	return f, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (f *File) Save(c context.Context) (*File, error) {
	// Update the Updated time
	f.Updated = time.Now()

	k, err := nds.Put(c, f.key(c, "File"), f)
	if err != nil {
		return nil, err
	}
	f.Id = k.IntID()
	return f, nil
}
