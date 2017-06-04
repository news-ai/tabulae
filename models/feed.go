package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	"github.com/qedus/nds"
)

type Feed struct {
	apiModels.Base

	FeedURL string `json:"url"`

	ContactId     int64 `json:"contactid" apiModel:"Contact"`
	ListId        int64 `json:"listid" apiModel:"MediaList"`
	PublicationId int64 `json:"publicationid" apiModel:"Publication"`

	ValidFeed bool `json:"validfeed"`
	Running   bool `json:"running"`
}

/*
* Private methods
 */

func (f *Feed) Key(c context.Context) *datastore.Key {
	return f.key(c, "Feed")
}

/*
* Create methods
 */

func (f *Feed) Create(c context.Context, r *http.Request, currentUser User) (*Feed, error) {
	f.CreatedBy = currentUser.Id
	f.Created = time.Now()

	// Initially the feed is both running and valid
	f.Running = true
	f.ValidFeed = true

	_, err := f.Save(c)
	return f, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (f *Feed) Save(c context.Context) (*Feed, error) {
	k, err := nds.Put(c, f.key(c, "Feed"), f)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	f.Id = k.IntID()
	return f, nil
}
