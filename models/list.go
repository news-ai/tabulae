package models

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type MediaList struct {
	Id int64 `json:"id" datastore:"-"`

	Contacts []Contact `json:"contacts"`

	CreatedBy User `json:"createdby" datastore:"-"`

	Created time.Time `json:"created"`
}

// Code to get data from App Engine:
// There should be double Lists. The name of the container is MediaList.
func defaultMediaListList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "MediaList", "default", 0, nil)
}

// Generates a new key for the data to be stored on App Engine
func (ml *MediaList) key(c appengine.Context) *datastore.Key {
	if ml.Id == 0 {
		ml.Created = time.Now()
		return datastore.NewIncompleteKey(c, "MediaList", nil)
	}
	return datastore.NewKey(c, "MediaList", "", 0, nil)
}

/*
* Get methods
 */

func getMediaList(c appengine.Context, id string) (MediaList, error) {
	// Get the MediaList by id
	mediaLists := []MediaList{}
	ks, err := datastore.NewQuery("MediaList").Filter("ID =", id).GetAll(c, &mediaLists)
	if err != nil {
		return MediaList{}, err
	}
	if len(mediaLists) > 0 {
		mediaLists[0].Id = ks[0].IntID()
		return mediaLists[0], nil
	}
	return MediaList{}, errors.New("No media list by this id")
}

/*
* Create methods
 */

func (ml *MediaList) create(c appengine.Context) (*MediaList, error) {
	currentUser, err := GetCurrentUser(c)
	if err != nil {
		return ml, err
	}

	ml.CreatedBy = currentUser
	ml.Created = time.Now()

	_, err = ml.save(c)
	return ml, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ml *MediaList) save(c appengine.Context) (*MediaList, error) {
	k, err := datastore.Put(c, ml.key(c), ml)
	if err != nil {
		return nil, err
	}
	ml.Id = k.IntID()
	return ml, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single media list
func GetMediaLists(c appengine.Context) ([]MediaList, error) {
	mediaLists := []MediaList{}
	ks, err := datastore.NewQuery("MediaList").GetAll(c, &mediaLists)
	if err != nil {
		return []MediaList{}, err
	}
	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Id = ks[i].IntID()
	}
	return mediaLists, nil
}

func GetMediaList(c appengine.Context, id string) (MediaList, error) {
	// Get the details of the current user
	mediaList, err := getMediaList(c, id)
	if err != nil {
		return MediaList{}, err
	}
	return mediaList, nil
}
