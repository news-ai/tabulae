package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	// "strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	// "google.golang.org/appengine/memcache"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/api/controllers"
	apiModels "github.com/news-ai/api/models"
	apiSearch "github.com/news-ai/api/search"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
)

/*
* Private
 */

var nonCustomHeaders = []string{"firstname", "lastname", "email", "employers", "pastemployers", "notes", "linkedin", "twitter", "instagram", "website", "blog", "phonenumber", "location"}
var nonCustomHeadersName = []string{"First Name", "Last Name", "Email", "Employers", "Past Employers", "Notes", "Linkedin", "Twitter", "Instagram", "Website", "Blog", "Phone #", "Location"}

var customHeaders = []string{"instagramfollowers", "instagramfollowing", "instagramlikes", "instagramcomments", "instagramposts", "twitterfollowers", "twitterfollowing", "twitterlikes", "twitterretweets", "twitterposts", "latestheadline", "lastcontacted", "publicationlastcontacted"}
var customHeadersName = []string{"Instagram Followers", "Instagram Following", "Instagram Likes", "Instagram Comments", "Instagram Posts", "Twitter Followers", "Twitter Following", "Twitter Likes", "Twitter Retweets", "Twitter Posts", "Latest Headline", "Last Contacted", "Publication Last Contacted"}

type duplicateListDetails struct {
	Name string `json:"name"`
}

/*
* Private methods
 */

/*
* Get methods
 */

func getMediaListBasic(c context.Context, r *http.Request, id int64) (models.MediaList, error) {
	if id == 0 {
		return models.MediaList{}, errors.New("datastore: no such entity")
	}

	// Get the MediaList by id
	var mediaList models.MediaList
	mediaListId := datastore.NewKey(c, "MediaList", "", id, nil)

	err := nds.Get(c, mediaListId, &mediaList)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	if !mediaList.Created.IsZero() {
		mediaList.Format(mediaListId, "lists")
		mediaList.AddNewCustomFieldsMapToOldLists(c)

		if !mediaList.PublicList {
			user, err := controllers.GetCurrentUser(c, r)
			if err != nil {
				log.Errorf(c, "%v", err)
				return models.MediaList{}, errors.New("Could not get user")
			}

			// 3 ways to check if user can access media list:
			// 1. If admin
			// 2. If created by user
			// 3. If is within the user's team
			if mediaList.CreatedBy != user.Id && !user.IsAdmin {
				if mediaList.TeamId == 0 || user.TeamId == 0 || mediaList.TeamId != user.TeamId {
					return models.MediaList{}, errors.New("Forbidden")
				}
			}
		}

		return mediaList, nil
	}
	return models.MediaList{}, errors.New("No media list by this id")
}

func getMediaList(c context.Context, r *http.Request, id int64) (models.MediaList, error) {
	if id == 0 {
		return models.MediaList{}, errors.New("datastore: no such entity")
	}

	// Get the MediaList by id
	var mediaList models.MediaList
	mediaListId := datastore.NewKey(c, "MediaList", "", id, nil)

	err := nds.Get(c, mediaListId, &mediaList)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	if !mediaList.Created.IsZero() {
		mediaList.Format(mediaListId, "lists")
		mediaList.AddNewCustomFieldsMapToOldLists(c)

		if !mediaList.PublicList {
			user, err := controllers.GetCurrentUser(c, r)
			if err != nil {
				log.Errorf(c, "%v", err)
				return models.MediaList{}, errors.New("Could not get user")
			}

			// 3 ways to check if user can access media list:
			// 1. If admin
			// 2. If created by user
			// 3. If is within the user's team
			if mediaList.CreatedBy != user.Id && !user.IsAdmin {
				if mediaList.TeamId == 0 || user.TeamId == 0 || mediaList.TeamId != user.TeamId {
					return models.MediaList{}, errors.New("Forbidden")
				}
			}

			// If it is empty but there are still contacts by this list then populate them
			// This is a data correction problem
			if len(mediaList.Contacts) == 0 {
				contacts, err := filterContactsForListId(c, r, mediaList.Id)
				if err != nil {
					return mediaList, nil
				}

				contactIds := []int64{}
				for i := 0; i < len(contacts); i++ {
					contactIds = append(contactIds, contacts[i].Id)
				}

				mediaList.Contacts = contactIds
				mediaList.Save(c)
			}
		}

		return mediaList, nil
	}
	return models.MediaList{}, errors.New("No media list by this id")
}

func getFieldsMap() []models.CustomFieldsMap {
	fieldsmap := []models.CustomFieldsMap{}

	for i := 0; i < len(nonCustomHeaders); i++ {
		field := models.CustomFieldsMap{
			Name:        nonCustomHeadersName[i],
			Value:       nonCustomHeaders[i],
			CustomField: false,
			Hidden:      false,
		}
		fieldsmap = append(fieldsmap, field)
	}

	for i := 0; i < len(customHeaders); i++ {
		field := models.CustomFieldsMap{
			Name:        customHeadersName[i],
			Value:       customHeaders[i],
			CustomField: true,
			Hidden:      true,
		}
		fieldsmap = append(fieldsmap, field)
	}

	return fieldsmap
}

func duplicateList(c context.Context, r *http.Request, id string, name string) (models.MediaList, interface{}, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	if mediaList.TeamId != user.TeamId && mediaList.CreatedBy != user.Id && !user.IsAdmin {
		return models.MediaList{}, nil, errors.New("Forbidden")
	}

	if name == "" {
		name = "Copy of " + mediaList.Name
	}

	previousContacts := mediaList.Contacts

	// Duplicate a list
	mediaList.Id = 0
	mediaList.Name = name
	mediaList.Contacts = []int64{}
	mediaList.PublicList = false
	mediaList.CreatedBy = user.Id
	mediaList.Create(c, r, user)

	contacts := []models.Contact{}
	for i := 0; i < len(previousContacts); i++ {
		contact, err := getContact(c, r, previousContacts[i])
		if err != nil {
			log.Errorf(c, "%v", err)
		} else {
			contact.ListId = 0
			contacts = append(contacts, contact)
		}
	}

	newContacts, err := BatchCreateContactsForDuplicateList(c, r, contacts, mediaList.Id)
	if err != nil {
		return models.MediaList{}, nil, err
	}

	mediaList.Contacts = newContacts
	mediaList.Save(c)

	sync.ListUploadResourceBulkSync(r, mediaList.Id, mediaList.Contacts, []int64{})
	return mediaList, nil, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single media list
func GetMediaLists(c context.Context, r *http.Request, archived bool) ([]models.MediaList, interface{}, int, int, error) {
	mediaLists := []models.MediaList{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	// If the user is active then we can return their media lists
	if user.IsActive {
		query := datastore.NewQuery("MediaList").Filter("CreatedBy =", user.Id).Filter("Archived =", archived)
		// if archived {
		// 	query = query.Filter("IsDeleted =", false)
		// }
		query = query.Filter("PublicList =", false)

		query = controllers.ConstructQuery(query, r)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.MediaList{}, nil, 0, 0, err
		}

		mediaLists = make([]models.MediaList, len(ks))
		err = nds.GetMulti(c, ks, mediaLists)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.MediaList{}, nil, 0, 0, err
		}

		for i := 0; i < len(mediaLists); i++ {
			mediaLists[i].Format(ks[i], "lists")
			mediaLists[i].AddNewCustomFieldsMapToOldLists(c)
		}

		// Go through their media lists and add TeamID if not present
		if user.TeamId != 0 {
			for i := 0; i < len(mediaLists); i++ {
				if mediaLists[i].TeamId == 0 {
					mediaLists[i].TeamId = user.TeamId
					mediaLists[i].Save(c)
				}
			}
		}

		queryField := gcontext.Get(r, "q").(string)
		if queryField != "" {
			fieldSelector := strings.Split(queryField, ":")
			if len(fieldSelector) != 2 {
				selectedLists, total, err := search.SearchListsByAll(c, r, queryField, user.Id)
				if err != nil {
					return nil, nil, 0, 0, err
				}

				selectedMediaLists := []models.MediaList{}
				for i := 0; i < len(selectedLists); i++ {
					singleMediaList, err := getMediaList(c, r, selectedLists[i].Id)
					if err == nil {
						selectedMediaLists = append(selectedMediaLists, singleMediaList)
					}
				}

				return selectedMediaLists, nil, len(selectedMediaLists), total, nil
			}

			if fieldSelector[0] == "client" || fieldSelector[0] == "tag" {
				selectedLists, total, err := search.SearchListsByFieldSelector(c, r, fieldSelector[0], fieldSelector[1], user.Id)
				if err != nil {
					return nil, nil, 0, 0, err
				}

				selectedMediaLists := []models.MediaList{}
				for i := 0; i < len(selectedLists); i++ {
					singleMediaList, err := getMediaList(c, r, selectedLists[i].Id)
					if err == nil {
						selectedMediaLists = append(selectedMediaLists, singleMediaList)
					}
				}

				return selectedMediaLists, nil, len(selectedMediaLists), total, nil
			}
		}
	}

	// If the user is not active then we block their media lists
	return mediaLists, nil, len(mediaLists), 0, nil
}

func GetMediaListsClients(c context.Context, r *http.Request) (interface{}, interface{}, int, int, error) {
	clients := struct {
		Clients []string `json:"clients"`
	}{
		[]string{},
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return clients, nil, 0, 0, err
	}

	query := datastore.NewQuery("MediaList").Filter("CreatedBy =", user.Id).Filter("Archived =", false)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return clients, nil, 0, 0, err
	}

	mediaLists := make([]models.MediaList, len(ks))
	err = nds.GetMulti(c, ks, mediaLists)
	if err != nil {
		log.Errorf(c, "%v", err)
		return clients, nil, 0, 0, err
	}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Format(ks[i], "lists")
	}

	uniqueClients := map[string]bool{}
	for i := 0; i < len(mediaLists); i++ {
		if mediaLists[i].Client != "" {
			uniqueClients[mediaLists[i].Client] = true
		}
	}

	keys := make([]string, 0, len(uniqueClients))
	for k := range uniqueClients {
		keys = append(keys, k)
	}

	clients.Clients = keys
	return clients, nil, len(clients.Clients), 0, nil
}

func GetAllMediaLists(c context.Context, r *http.Request) ([]models.MediaList, error) {
	query := datastore.NewQuery("MediaList")
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, err
	}

	mediaLists := make([]models.MediaList, len(ks))
	err = nds.GetMulti(c, ks, mediaLists)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, err
	}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Format(ks[i], "lists")
	}

	return mediaLists, nil
}

// Gets every single media list
func GetPublicMediaLists(c context.Context, r *http.Request) ([]models.MediaList, interface{}, int, int, error) {
	mediaLists := []models.MediaList{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	query := datastore.NewQuery("MediaList").Filter("PublicList =", true).Filter("Archived =", false)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	mediaLists = make([]models.MediaList, len(ks))
	err = nds.GetMulti(c, ks, mediaLists)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Format(ks[i], "lists")

		if mediaLists[i].PublicList && user.Id != mediaLists[i].CreatedBy {
			mediaLists[i].ReadOnly = true
		}
	}

	return mediaLists, nil, len(mediaLists), 0, nil
}

// Gets all of the team media lists
// Excludes any media list that is logged in users
func GetTeamMediaLists(c context.Context, r *http.Request) ([]models.MediaList, interface{}, int, int, error) {
	mediaLists := []models.MediaList{}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	if user.TeamId == 0 {
		return []models.MediaList{}, nil, 0, 0, errors.New("You are not a part of a team")
	}

	query := datastore.NewQuery("MediaList").Filter("TeamId =", user.TeamId).Filter("Archived =", false)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	mediaLists = make([]models.MediaList, len(ks))
	err = nds.GetMulti(c, ks, mediaLists)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, 0, err
	}

	mediaListsOthers := []models.MediaList{}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Format(ks[i], "lists")
		if mediaLists[i].CreatedBy != user.Id {
			mediaListsOthers = append(mediaListsOthers, mediaLists[i])
		}
	}

	return mediaListsOthers, nil, len(mediaListsOthers), 0, nil
}

func GetMediaList(c context.Context, r *http.Request, id string) (models.MediaList, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	mediaList, err := getMediaList(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	if mediaList.PublicList {
		mediaList.ReadOnly = true
	}

	return mediaList, nil, nil
}

/*
* Create methods
 */

func CreateMediaList(c context.Context, w http.ResponseWriter, r *http.Request) (models.MediaList, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var medialist models.MediaList
	err := decoder.Decode(buf, &medialist)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return medialist, nil, err
	}

	// Initial values for fieldsmap
	if len(medialist.FieldsMap) > 0 {
		medialist.FieldsMap = append(getFieldsMap(), medialist.FieldsMap...)
	} else {
		medialist.FieldsMap = getFieldsMap()
	}

	medialist.TeamId = currentUser.TeamId

	// Create media list
	_, err = medialist.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	sync.ResourceSync(r, medialist.Id, "List", "create")
	return medialist, nil, nil
}

func CreateSampleMediaList(c context.Context, r *http.Request, user apiModels.User) (models.MediaList, interface{}, error) {
	// Create a fake media list
	mediaList := models.MediaList{}
	mediaList.Name = "My first list!"
	mediaList.Client = "Microsoft"
	mediaList.FieldsMap = getFieldsMap()

	field := models.CustomFieldsMap{
		Name:        "This is a custom column",
		Value:       "This is a custom column",
		CustomField: true,
		Hidden:      false,
	}
	mediaList.FieldsMap = append(mediaList.FieldsMap, field)

	mediaList.CreatedBy = user.Id
	mediaList.Created = time.Now()
	mediaList.Save(c)

	// Create a new contact for this list
	contacts := []int64{}
	singleContact := models.Contact{}
	singleContact.FirstName = "Shereen"
	singleContact.LastName = "Bhan"
	singleContact.Email = "shereen.bhan@network18online.com"
	singleContact.Twitter = "shereenbhan"
	singleContact.Website = "http://www.moneycontrol.com/cnbctv18/"
	singleContact.CreatedBy = user.Id
	singleContact.Employers = []int64{6399756150505472}
	singleContact.Created = time.Now()
	singleContact.ListId = mediaList.Id
	_, err := createContact(c, r, &singleContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return mediaList, nil, err
	}

	// Add a contact into the list
	contacts = append(contacts, singleContact.Id)

	fashionContact := models.Contact{}
	fashionContact.FirstName = "Chiara"
	fashionContact.LastName = "Ferragni"
	fashionContact.Email = "contact@tbscrew.com"
	fashionContact.Twitter = "chiaraferragni"
	fashionContact.Instagram = "chiaraferragni"
	fashionContact.LinkedIn = "https://www.linkedin.com/in/chiara-ferragni-2b4262101"
	fashionContact.Blog = "http://www.theblondesalad.com/"
	fashionContact.Website = "http://www.theblondesalad.com/"
	fashionContact.CreatedBy = user.Id
	fashionContact.Employers = []int64{5308689770610688}
	fashionContact.Created = time.Now()
	fashionContact.ListId = mediaList.Id

	customField := models.CustomContactField{}
	customField.Name = "This is a custom column"
	customField.Value = "This is a custom value"

	fashionContact.CustomFields = append(fashionContact.CustomFields, customField)
	_, err = createContact(c, r, &fashionContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return mediaList, nil, err
	}

	// Add a contact into the list
	contacts = append(contacts, fashionContact.Id)

	mediaList.Contacts = contacts
	mediaList.Save(c)

	// Create a fake feed
	feed := models.Feed{}
	feed.FeedURL = "http://www.firstpost.com/tag/shereen-bhan/feed"
	feed.ContactId = singleContact.Id
	feed.ListId = mediaList.Id
	feed.PublicationId = 5594198795354112
	feed.Create(c, r, user)

	// Create a fake feed
	fashionFeed := models.Feed{}
	fashionFeed.FeedURL = "http://www.theblondesalad.com/feed"
	fashionFeed.ContactId = fashionContact.Id
	fashionFeed.ListId = mediaList.Id
	fashionFeed.PublicationId = 5308689770610688
	fashionFeed.Create(c, r, user)

	sync.ResourceSync(r, mediaList.Id, "List", "create")
	return mediaList, nil, nil
}

/*
* Update methods
 */

func UpdateMediaList(c context.Context, r *http.Request, id string) (models.MediaList, interface{}, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	if mediaList.TeamId != user.TeamId && mediaList.CreatedBy != user.Id && !user.IsAdmin {
		return models.MediaList{}, nil, errors.New("Forbidden")
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedMediaList models.MediaList
	err = decoder.Decode(buf, &updatedMediaList)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	if len(updatedMediaList.Contacts) > 0 {
		mediaList.Contacts = updatedMediaList.Contacts
	} else {
		if mediaList.Name == updatedMediaList.Name {
			if len(updatedMediaList.FieldsMap) == 0 {
				utilities.UpdateIfNotBlank(&mediaList.Client, updatedMediaList.Client)
				if len(updatedMediaList.Tags) > 0 {
					mediaList.Tags = updatedMediaList.Tags
				}

				// If you want to empty a list
				if len(mediaList.Tags) > 0 && len(updatedMediaList.Tags) == 0 {
					mediaList.Tags = updatedMediaList.Tags
				}
			}
		} else {
			utilities.UpdateIfNotBlank(&mediaList.Name, updatedMediaList.Name)
		}
	}

	// Edge case for when you want to empty the list & there's only 1 contact
	if len(mediaList.Contacts) == 1 {
		// Get the single contact that the mediaList has
		singleContact, err := getContact(c, r, mediaList.Contacts[0])
		if err == nil {
			// If the singleContact has been deleted then we set the mediaList
			// contacts to empty
			if singleContact.IsDeleted {
				mediaList.Contacts = []int64{}
			}
		}
	}

	// Edge case for when you want to empty the list
	contactsInList, err := filterContactsForListId(c, r, mediaList.Id)
	if len(contactsInList) == 0 {
		mediaList.Contacts = []int64{}
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

	// If new media list wants to be subscribed to then subscribe to it
	if updatedMediaList.Subscribed == true {
		mediaList.Subscribed = true
	}

	if mediaList.Subscribed == true && updatedMediaList.Subscribed == false {
		mediaList.Subscribed = false
	}

	_, mediaListSaveErr := mediaList.Save(c)

	// If there's a problem saving the document
	if mediaListSaveErr != nil {
		log.Errorf(c, "%v", err)
		mediaList.Save(c)
	}
	sync.ResourceSync(r, mediaList.Id, "List", "create")
	return mediaList, nil, nil
}

func UpdateMediaListToPublic(c context.Context, r *http.Request, id string) (models.MediaList, interface{}, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}
	if !user.IsAdmin {
		return models.MediaList{}, nil, errors.New("Forbidden")
	}

	mediaList.PublicList = !mediaList.PublicList

	mediaList.Save(c)
	sync.ResourceSync(r, mediaList.Id, "List", "create")
	return mediaList, nil, nil
}

func ReSyncMediaList(c context.Context, r *http.Request, id string) (models.MediaList, interface{}, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}
	if !user.IsAdmin {
		return models.MediaList{}, nil, errors.New("Forbidden")
	}

	sync.ListUploadResourceBulkSync(r, mediaList.Id, mediaList.Contacts, []int64{})
	sync.ResourceSync(r, mediaList.Id, "List", "create")
	return mediaList, nil, nil
}

/*
* Action methods
 */

func GetContactsForList(c context.Context, r *http.Request, id string) ([]models.Contact, interface{}, int, int, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	queryField := gcontext.Get(r, "q").(string)
	if queryField != "" {
		contacts, total, err := search.SearchContactsByList(c, r, queryField, user, mediaList.CreatedBy, mediaList.Id)
		if err != nil {
			return []models.Contact{}, nil, 0, 0, err
		}

		publications := contactsToPublications(c, contacts)
		return contacts, publications, len(contacts), total, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	startPosition := offset
	endPosition := startPosition + limit

	if len(mediaList.Contacts) < startPosition {
		return []models.Contact{}, nil, 0, 0, err
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

	// Check memcache of this cluster of contacts are in memcache
	// Current case is if limit is 50 (as a prototype) so we don't
	// have to look for expiration keys
	// if limit == 50 {
	// 	memcacheKey := id + "_" + strconv.Itoa(startPosition) + "_" + strconv.Itoa(endPosition)
	// 	_, err := memcache.Gob.Get(c, memcacheKey, &contacts)

	// 	if err != nil {
	// 		err = nds.GetMulti(c, subsetKeyIds, contacts)
	// 		if err != nil {
	// 			log.Errorf(c, "%v", err)
	// 			return []models.Contact{}, nil, 0, 0, err
	// 		}

	// 		for i := 0; i < len(contacts); i++ {
	// 			if contacts[i].ListId == 0 {
	// 				contacts[i].ListId = mediaList.Id
	// 				contacts[i].Save(c, r)
	// 			}

	// 			contacts[i].Id = subsetIds[i]
	// 			contacts[i].Type = "contacts"
	// 		}

	// 		contacts, err = ContactsToDefaultFields(c, r, contacts, mediaList)
	// 		if err != nil {
	// 			log.Errorf(c, "%v", err)
	// 		}

	// 		item1 := &memcache.Item{
	// 			Key:    memcacheKey,
	// 			Object: &contacts,
	// 		}
	// 		err = memcache.Gob.Set(c, item1)
	// 		if err != nil {
	// 			log.Errorf(c, "%v", err)
	// 		}
	// 	} else {
	// 		log.Infof(c, "memcache: %v", contacts)
	// 	}

	// 	publications := contactsToPublications(c, contacts)
	// 	return contacts, publications, len(contacts), len(mediaList.Contacts), nil
	// }

	err = nds.GetMulti(c, subsetKeyIds, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	for i := 0; i < len(contacts); i++ {
		if contacts[i].ListId == 0 {
			contacts[i].ListId = mediaList.Id
			contacts[i].Save(c, r)
		}

		contacts[i].Id = subsetIds[i]
		contacts[i].Type = "contacts"
	}

	contacts, err = ContactsToDefaultFields(c, r, contacts, mediaList)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	// Add includes
	publications := contactsToPublications(c, contacts)
	return contacts, publications, len(contacts), len(mediaList.Contacts), nil
}

func GetEmailsForList(c context.Context, r *http.Request, id string) ([]models.Email, interface{}, int, int, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails, count, err := filterEmailbyListId(c, r, mediaList.Id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Add includes
	mediaLists := emailsToLists(c, r, emails)
	contacts := emailsToContacts(c, r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, count, 0, nil
}

func GetHeadlinesForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ListId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	headlines, total, err := apiSearch.SearchHeadlinesByResourceId(c, r, feeds, []string{})
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return headlines, nil, len(headlines), total, nil
}

func GetFeedForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	mediaList, err := getMediaList(c, r, currentId)
	contactIds := []*datastore.Key{}
	for i := 0; i < len(mediaList.Contacts); i++ {
		contactIds = append(contactIds, datastore.NewKey(c, "Contact", "", mediaList.Contacts[i], nil))
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(mediaList.Contacts))

	err = nds.GetMulti(c, contactIds, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ListId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	feed, total, err := apiSearch.SearchFeedForContacts(c, r, contacts, feeds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return feed, nil, len(feed), total, nil
}

func GetTweetsForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	mediaList, err := getMediaList(c, r, currentId)
	contactIds := []*datastore.Key{}
	for i := 0; i < len(mediaList.Contacts); i++ {
		contactIds = append(contactIds, datastore.NewKey(c, "Contact", "", mediaList.Contacts[i], nil))
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(mediaList.Contacts))

	err = nds.GetMulti(c, contactIds, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	usernames := []string{}
	for i := 0; i < len(contacts); i++ {
		if contacts[i].Twitter != "" {
			usernames = append(usernames, contacts[i].Twitter)
		}
	}

	tweets, total, err := apiSearch.SearchTweetsByUsernames(c, r, usernames)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return tweets, nil, len(tweets), total, nil
}

func GetTwitterTimeseriesForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var contactIds models.ContactIdsArray
	err := decoder.Decode(buf, &contactIds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	twitterUsernames := []string{}

	for i := 0; i < len(contactIds.ContactIds); i++ {
		contact, err := getContact(c, r, contactIds.ContactIds[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, err
		}

		if contact.Twitter != "" {
			twitterUsernames = append(twitterUsernames, contact.Twitter)
		}
	}

	defaultDate := 7
	if contactIds.Days != 0 {
		defaultDate = contactIds.Days
	}

	twitterTimeseries, err := apiSearch.SearchTwitterTimeseriesByUsernamesWithDays(c, r, twitterUsernames, defaultDate)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return twitterTimeseries, nil, nil
}

func GetInstagramTimeseriesForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var contactIds models.ContactIdsArray
	err := decoder.Decode(buf, &contactIds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	instagramUsernames := []string{}

	for i := 0; i < len(contactIds.ContactIds); i++ {
		contact, err := getContact(c, r, contactIds.ContactIds[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, err
		}

		if contact.Instagram != "" {
			instagramUsernames = append(instagramUsernames, contact.Instagram)
		}
	}

	defaultDate := 7
	if contactIds.Days != 0 {
		defaultDate = contactIds.Days
	}

	instagramTimeseries, err := apiSearch.SearchInstagramTimeseriesByUsernamesWithDays(c, r, instagramUsernames, defaultDate)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return instagramTimeseries, nil, nil
}

func DuplicateList(c context.Context, r *http.Request, id string) (models.MediaList, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var duplicateDetails duplicateListDetails
	err := decoder.Decode(buf, &duplicateDetails)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	return duplicateList(c, r, id, duplicateDetails.Name)
}

func DeleteMediaList(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	mediaList, err := getMediaList(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	// Double check permissions. Admins should not be able to delete.
	if mediaList.TeamId != user.TeamId && !permissions.AccessToObject(mediaList.CreatedBy, user.Id) {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contactIds := []int64{}
	listId := mediaList.Id

	contacts, err := GetContactsByIds(c, r, mediaList.Contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	// Delete contacts
	keys := []*datastore.Key{}
	for i := 0; i < len(contacts); i++ {
		if contacts[i].ListId != 0 {
			listId = contacts[i].ListId
		}

		contacts[i].IsDeleted = true
		contacts[i].Save(c, r)

		keys = append(keys, contacts[i].Key(c))
		contactIds = append(contactIds, contacts[i].Id)
	}

	ks := []*datastore.Key{}
	err = nds.RunInTransaction(c, func(ctx context.Context) error {
		contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
		ks, err = nds.PutMulti(contextWithTimeout, keys, contacts)
		if err != nil {
			log.Errorf(c, "%v", err)
			return err
		}
		return nil
	}, nil)

	// Pubsub to sync listid and contactids
	sync.ListUploadResourceBulkSync(r, listId, contactIds, []int64{})

	// Pubsub to remove ES contact
	sync.ResourceSync(r, mediaList.Id, "List", "delete")

	_, err = mediaList.Delete(c)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return nil, nil, nil
}
