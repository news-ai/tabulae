package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

/*
* Private
 */

var nonCustomHeaders = []string{"firstname", "lastname", "email", "employers", "pastemployers", "notes", "linkedin", "twitter", "instagram", "website", "blog"}

/*
* Private methods
 */

/*
* Get methods
 */

func getMediaList(c context.Context, r *http.Request, id int64) (models.MediaList, error) {
	if id == 0 {
		return models.MediaList{}, errors.New("datastore: no such entity")
	}

	// Get the MediaList by id
	var mediaList models.MediaList
	mediaListId := datastore.NewKey(c, "MediaList", "", id, nil)

	err := nds.Get(c, mediaListId, &mediaList)
	if err != nil {
		return models.MediaList{}, err
	}

	if !mediaList.Created.IsZero() {
		mediaList.Id = mediaListId.IntID()

		user, err := GetCurrentUser(c, r)
		if err != nil {
			return models.MediaList{}, errors.New("Could not get user")
		}
		if mediaList.CreatedBy != user.Id {
			return models.MediaList{}, errors.New("Forbidden")
		}

		return mediaList, nil
	}
	return models.MediaList{}, errors.New("No media list by this id")
}

func getFieldsMap() []models.CustomFieldsMap {
	fieldsmap := []models.CustomFieldsMap{}

	for i := 0; i < len(nonCustomHeaders); i++ {
		field := models.CustomFieldsMap{
			Name:        nonCustomHeaders[i],
			Value:       nonCustomHeaders[i],
			CustomField: false,
			Hidden:      false,
		}
		fieldsmap = append(fieldsmap, field)
	}

	return fieldsmap
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single media list
func GetMediaLists(c context.Context, r *http.Request) ([]models.MediaList, error) {
	mediaLists := []models.MediaList{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.MediaList{}, err
	}

	ks, err := datastore.NewQuery("MediaList").Filter("CreatedBy =", user.Id).GetAll(c, &mediaLists)
	if err != nil {
		return []models.MediaList{}, err
	}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Id = ks[i].IntID()
	}
	return mediaLists, nil
}

func GetMediaList(c context.Context, r *http.Request, id string) (models.MediaList, error) {
	// Get the details of the current user
	currentId, err := utils.StringIdToInt(id)
	if err != nil {
		return models.MediaList{}, err
	}

	mediaList, err := getMediaList(c, r, currentId)
	if err != nil {
		return models.MediaList{}, err
	}
	return mediaList, nil
}

/*
* Create methods
 */

func CreateMediaList(c context.Context, w http.ResponseWriter, r *http.Request) (models.MediaList, error) {
	decoder := json.NewDecoder(r.Body)
	var medialist models.MediaList
	err := decoder.Decode(&medialist)
	if err != nil {
		return models.MediaList{}, err
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return medialist, err
	}

	// Initial values for fieldsmap
	medialist.FieldsMap = getFieldsMap()

	// Create media list
	_, err = medialist.Create(c, r, currentUser)
	if err != nil {
		return models.MediaList{}, err
	}

	return medialist, nil
}

/*
* Update methods
 */

func UpdateMediaList(c context.Context, r *http.Request, id string) (models.MediaList, error) {
	// Get the details of the current media list
	mediaList, err := GetMediaList(c, r, id)
	if err != nil {
		return models.MediaList{}, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return models.MediaList{}, err
	}
	if mediaList.CreatedBy != user.Id {
		return models.MediaList{}, errors.New("Forbidden")
	}

	decoder := json.NewDecoder(r.Body)
	var updatedMediaList models.MediaList
	err = decoder.Decode(&updatedMediaList)
	if err != nil {
		return models.MediaList{}, err
	}

	utils.UpdateIfNotBlank(&mediaList.Name, updatedMediaList.Name)

	if len(updatedMediaList.Contacts) > 0 {
		mediaList.Contacts = updatedMediaList.Contacts
	}
	if len(updatedMediaList.FieldsMap) > 0 {
		mediaList.FieldsMap = updatedMediaList.FieldsMap
	}

	// If new media list wants to be archived then archive it
	if updatedMediaList.Archived == true {
		mediaList.Archived = true
	}

	// If they are already archived and you want to unarchive the media list
	if mediaList.Archived == true && updatedMediaList.Archived == false {
		mediaList.Archived = false
	}

	mediaList.Save(c)
	return mediaList, nil
}

/*
* Action methods
 */

func GetContactsForList(c context.Context, r *http.Request, id string) (models.BaseResponse, error) {
	response := models.BaseResponse{}

	// Get the details of the current media list
	mediaList, err := GetMediaList(c, r, id)
	if err != nil {
		return response, err
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	log.Infof(c, "%v", offset)

	startPosition := offset
	endPosition := startPosition + limit

	if len(mediaList.Contacts) < startPosition {
		return response, nil
	}

	if len(mediaList.Contacts) < endPosition {
		endPosition = len(mediaList.Contacts)
	}

	subsetIds := mediaList.Contacts[startPosition:endPosition]
	subsetKeyIds := []*datastore.Key{}
	for i := 0; i < len(subsetIds); i++ {
		subsetKeyIds = append(subsetKeyIds, datastore.NewKey(c, "Contact", "", subsetIds[i], nil))
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(subsetIds))

	err = nds.GetMulti(c, subsetKeyIds, contacts)
	if err != nil {
		return response, err
	}

	publicationIds := []int64{}
	// externalResourceIds := map[string][]int64{}
	// externalResourceIds = make(map[string][]int64)

	for i := 0; i < len(contacts); i++ {
		if contacts[i].LinkedIn != "" {
			findOrCreateMasterContact(c, &contacts[i], r)
			linkedInSync(c, r, &contacts[i], false)
			checkAgainstParent(c, r, &contacts[i])
		}

		contacts[i].Id = subsetKeyIds[i].IntID()
		publicationIds = append(publicationIds, contacts[i].Employers...)
		publicationIds = append(publicationIds, contacts[i].PastEmployers...)
	}

	response.Count = len(contacts)
	response.Results = contacts

	// Work on includes
	publications := []models.Publication{}
	publicationExists := map[int64]bool{}
	publicationExists = make(map[int64]bool)

	for i := 0; i < len(publicationIds); i++ {
		if _, ok := publicationExists[publicationIds[i]]; !ok {
			publication, _ := getPublication(c, publicationIds[i])
			publications = append(publications, publication)
			publicationExists[publicationIds[i]] = true
		}
	}

	response.Includes = publications

	return response, nil
}
