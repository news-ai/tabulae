package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type MediaList struct {
	Id int64 `json:"id" datastore:"-"`

	Name string `json:"name"`

	Contacts     []int64  `json:"contacts"`
	CustomFields []string `json:"customfields"`

	FileUpload int64 `json:"fileupload"`

	Archived bool `json:"archived"`

	CreatedBy int64 `json:"createdby"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// Generates a new key for the data to be stored on App Engine
func (ml *MediaList) key(c appengine.Context) *datastore.Key {
	if ml.Id == 0 {
		ml.Created = time.Now()
		return datastore.NewIncompleteKey(c, "MediaList", nil)
	}
	return datastore.NewKey(c, "MediaList", "", ml.Id, nil)
}

/*
* Get methods
 */

func getMediaList(c appengine.Context, r *http.Request, id int64) (MediaList, error) {
	// Get the MediaList by id
	mediaLists := []MediaList{}
	mediaListId := datastore.NewKey(c, "MediaList", "", id, nil)
	ks, err := datastore.NewQuery("MediaList").Filter("__key__ =", mediaListId).GetAll(c, &mediaLists)
	if err != nil {
		return MediaList{}, err
	}
	if len(mediaLists) > 0 {
		mediaLists[0].Id = ks[0].IntID()

		user, err := GetCurrentUser(c, r)
		if err != nil {
			return MediaList{}, errors.New("Could not get user")
		}
		if mediaLists[0].CreatedBy != user.Id {
			return MediaList{}, errors.New("Forbidden")
		}

		return mediaLists[0], nil
	}
	return MediaList{}, errors.New("No media list by this id")
}

/*
* Create methods
 */

func (ml *MediaList) create(c appengine.Context, r *http.Request) (*MediaList, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return ml, err
	}

	ml.CreatedBy = currentUser.Id
	ml.Created = time.Now()

	_, err = ml.save(c)
	return ml, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ml *MediaList) save(c appengine.Context) (*MediaList, error) {
	// Update the Updated time
	ml.Updated = time.Now()

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
func GetMediaLists(c appengine.Context, r *http.Request) ([]MediaList, error) {
	mediaLists := []MediaList{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []MediaList{}, err
	}

	ks, err := datastore.NewQuery("MediaList").Filter("CreatedBy =", user.Id).GetAll(c, &mediaLists)
	if err != nil {
		return []MediaList{}, err
	}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Id = ks[i].IntID()
	}
	return mediaLists, nil
}

func GetMediaList(c appengine.Context, r *http.Request, id string) (MediaList, error) {
	// Get the details of the current user
	currentId, err := StringIdToInt(id)
	if err != nil {
		return MediaList{}, err
	}

	mediaList, err := getMediaList(c, r, currentId)
	if err != nil {
		return MediaList{}, err
	}
	return mediaList, nil
}

/*
* Create methods
 */

func CreateMediaList(c appengine.Context, w http.ResponseWriter, r *http.Request) (MediaList, error) {
	decoder := json.NewDecoder(r.Body)
	var medialist MediaList
	err := decoder.Decode(&medialist)
	if err != nil {
		return MediaList{}, err
	}

	// Create media list
	_, err = medialist.create(c, r)
	if err != nil {
		return MediaList{}, err
	}

	return medialist, nil
}

/*
* Update methods
 */

func UpdateMediaList(c appengine.Context, r *http.Request, id string) (MediaList, error) {
	// Get the details of the current media list
	mediaList, err := GetMediaList(c, r, id)
	if err != nil {
		return MediaList{}, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return MediaList{}, err
	}
	if mediaList.CreatedBy != user.Id {
		return MediaList{}, errors.New("Forbidden")
	}

	decoder := json.NewDecoder(r.Body)
	var updatedMediaList MediaList
	err = decoder.Decode(&updatedMediaList)
	if err != nil {
		return MediaList{}, err
	}

	UpdateIfNotBlank(&mediaList.Name, updatedMediaList.Name)
	if len(updatedMediaList.Contacts) > 0 {
		mediaList.Contacts = updatedMediaList.Contacts
	}
	if len(updatedMediaList.CustomFields) > 0 {
		mediaList.CustomFields = updatedMediaList.CustomFields
	}

	// If new media list wants to be archived then archive it
	if updatedMediaList.Archived == true {
		mediaList.Archived = true
	}

	// If they are already archived and you want to unarchive the media list
	if mediaList.Archived == true && updatedMediaList.Archived == false {
		mediaList.Archived = false
	}

	mediaList.save(c)
	return mediaList, nil
}
