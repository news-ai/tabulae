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

	Contacts []int64 `json:"contacts"`

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

func getMediaList(c appengine.Context, id int64) (MediaList, error) {
	// Get the MediaList by id
	mediaLists := []MediaList{}
	mediaListId := datastore.NewKey(c, "MediaList", "", id, nil)
	ks, err := datastore.NewQuery("MediaList").Filter("__key__ =", mediaListId).GetAll(c, &mediaLists)
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
	currentId, err := StringIdToInt(id)
	if err != nil {
		return MediaList{}, err
	}

	mediaList, err := getMediaList(c, currentId)
	if err != nil {
		return MediaList{}, err
	}
	return mediaList, nil
}

/*
* Create methods
 */

// Method not completed
func CreateMediaList(c appengine.Context, w http.ResponseWriter, r *http.Request) (MediaList, error) {
	decoder := json.NewDecoder(r.Body)
	var medialist MediaList
	err := decoder.Decode(&medialist)
	if err != nil {
		return MediaList{}, err
	}

	// Contacts in Media List
	for i := 0; i < len(medialist.Contacts); i++ {
		contact, err := getContact(c, medialist.Contacts[i])
		if err != nil {
			return MediaList{}, err
		}
		medialist.MediaContacts = append(medialist.MediaContacts, contact)
	}

	// Create contact
	_, err = medialist.create(c)
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
	mediaList, err := GetMediaList(c, id)
	if err != nil {
		return MediaList{}, err
	}

	decoder := json.NewDecoder(r.Body)
	var updatedMediaList MediaList
	err = decoder.Decode(&updatedMediaList)
	if err != nil {
		return MediaList{}, err
	}

	mediaList.Name = updatedMediaList.Name

	// Media List Contacts
	newMediaListMediaContacts := []Contact{}
	newMediaListContacts := []int64{}
	for i := 0; i < len(updatedMediaList.Contacts); i++ {
		contact, err := getContact(c, updatedMediaList.Contacts[i])
		if err != nil {
			return MediaList{}, err
		}
		newMediaListContacts = append(newMediaListContacts, contact.Id)
		newMediaListMediaContacts = append(newMediaListMediaContacts, contact)
	}
	if len(newMediaListContacts) > 0 {
		mediaList.Contacts = newMediaListContacts
		mediaList.MediaContacts = newMediaListMediaContacts
	}

	mediaList.save(c)
	return mediaList, nil
}
