package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"

	"github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/models"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getContactV2(c context.Context, r *http.Request, id int64) (models.ContactV2, error) {
	if id == 0 {
		return models.ContactV2{}, errors.New("datastore: no such entity")
	}
	// Get the ContactV2 by id
	var contact models.ContactV2
	contactId := datastore.NewKey(c, "ContactV2", "", id, nil)
	err := nds.Get(c, contactId, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.ContactV2{}, err
	}

	if !contact.Created.IsZero() {
		contact.Format(contactId, "contacts_v2")

		user, err := controllers.GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.ContactV2{}, err
		}

		if !permissions.AccessToObject(contact.TeamId, user.TeamId) && !user.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.ContactV2{}, err
		}

		return contact, nil
	}

	return models.ContactV2{}, errors.New("No contact by this id")
}

func contactsV2ToPublications(c context.Context, contacts []models.ContactV2) []models.Publication {
	publicationIds := []int64{}

	for i := 0; i < len(contacts); i++ {
		publicationIds = append(publicationIds, contacts[i].Employers...)
		publicationIds = append(publicationIds, contacts[i].PastEmployers...)
	}

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

	return publications
}

func contactsV2ToLists(c context.Context, r *http.Request, contacts []models.ContactV2) []models.MediaList {
	mediaListIds := []int64{}

	for i := 0; i < len(contacts); i++ {
		mediaListIds = append(mediaListIds, contacts[i].ListIds...)
	}

	// Work on includes
	mediaLists := []models.MediaList{}
	mediaListExists := map[int64]bool{}
	mediaListExists = make(map[int64]bool)

	for i := 0; i < len(mediaListIds); i++ {
		if _, ok := mediaListExists[mediaListIds[i]]; !ok {
			mediaList, _ := getMediaList(c, r, mediaListIds[i])
			mediaLists = append(mediaLists, mediaList)
			mediaListExists[mediaListIds[i]] = true
		}
	}

	return mediaLists
}

func getIncludesForContactsV2(c context.Context, r *http.Request, contacts []models.ContactV2) interface{} {
	mediaLists := contactsV2ToLists(c, r, contacts)
	publications := contactsV2ToPublications(c, contacts)

	includes := make([]interface{}, len(mediaLists)+len(publications))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(publications); i++ {
		includes[i+len(mediaLists)] = publications[i]
	}

	return includes
}

/*
* Update methods
 */

func updateContactV2(c context.Context, r *http.Request, contact *models.ContactV2, updatedContact models.ContactV2) (models.ContactV2, interface{}, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return *contact, nil, err
	}

	return *contact, nil, err
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single contact
func GetContactsV2(c context.Context, r *http.Request) ([]models.ContactV2, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.ContactV2{}, nil, 0, 0, err
	}

	// If the user is currently active
	if user.IsActive {
		query := datastore.NewQuery("ContactV2").Filter("TeamId =", user.TeamId)
		query = controllers.ConstructQuery(query, r)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.ContactV2{}, nil, 0, 0, err
		}

		contacts := []models.ContactV2{}
		contacts = make([]models.ContactV2, len(ks))
		err = nds.GetMulti(c, ks, contacts)
		if err != nil {
			log.Errorf(c, "%v", err)
			return contacts, nil, 0, 0, err
		}

		for i := 0; i < len(contacts); i++ {
			contacts[i].Format(ks[i], "contacts_v2")
		}

		includes := getIncludesForContactsV2(c, r, contacts)
		return contacts, nil, len(contacts), 0, nil
	}

	// If user is not active then return empty lists
	return []models.ContactV2{}, nil, 0, 0, nil
}

func GetContactV2(c context.Context, r *http.Request, id string) (models.ContactV2, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.ContactV2{}, nil, err
	}

	contact, err := getContactV2(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.ContactV2{}, nil, err
	}

	includes := getIncludesForContactsV2(c, r, []models.ContactV2{contact})
	return contact, nil, nil
}
