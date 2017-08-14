package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	"github.com/qedus/nds"
)

type File struct {
	apiModels.Base

	OriginalName string `json:"originalname"`
	FileName     string `json:"filename"`
	ListId       int64  `json:"listid" apiModel:"MediaList"`
	EmailId      int64  `json:"emailid" apiModel:"Email"`

	Url string `json:"url"`

	HeaderNames []string `json:"headernames" datastore:",noindex"`
	Order       []string `json:"order" datastore:",noindex"`

	Imported   bool `json:"imported"`
	FileExists bool `json:"fileexists"`
}

type FileOrder struct {
	HeaderNames []string `json:"headernames"`
	Order       []string `json:"order"`
	Sheet       string   `json:"string"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (f *File) Create(c context.Context, r *http.Request, currentUser apiModels.User) (*File, error) {
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

	k, err := nds.Put(c, f.BaseKey(c, "File"), f)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	f.Id = k.IntID()
	return f, nil
}
