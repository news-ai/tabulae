package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/utilities"
)

/*
* Private
 */

var nonCustomHeaders = []string{"firstname", "lastname", "email", "employers", "pastemployers", "notes", "linkedin", "twitter", "instagram", "website", "blog"}

var customHeaders = []string{"instagramfollowers"}

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
		log.Errorf(c, "%v", err)
		return models.MediaList{}, err
	}

	if !mediaList.Created.IsZero() {
		mediaList.Format(mediaListId, "lists")
		mediaList.AddNewFieldsMapToOldLists(c)

		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.MediaList{}, errors.New("Could not get user")
		}
		if mediaList.CreatedBy != user.Id && !user.IsAdmin {
			return models.MediaList{}, errors.New("Forbidden")
		}

		return mediaList, nil
	}
	return models.MediaList{}, errors.New("No media list by this id")
}

func getFieldsMap() []models.CustomFieldsMap {
	fieldsmap := []models.CustomFieldsMap{}

	for i := 0; i < len(nonCustomHeaders); i++ {
		isHidden := false

		if nonCustomHeaders[i] == "employers" || nonCustomHeaders[i] == "pastemployers" || nonCustomHeaders[i] == "instagramfollowers" {
			isHidden = true
		}

		field := models.CustomFieldsMap{
			Name:        nonCustomHeaders[i],
			Value:       nonCustomHeaders[i],
			CustomField: false,
			Hidden:      isHidden,
		}
		fieldsmap = append(fieldsmap, field)
	}

	for i := 0; i < len(customHeaders); i++ {
		field := models.CustomFieldsMap{
			Name:        customHeaders[i],
			Value:       customHeaders[i],
			CustomField: true,
			Hidden:      true,
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
func GetMediaLists(c context.Context, r *http.Request, archived bool) ([]models.MediaList, interface{}, int, error) {
	mediaLists := []models.MediaList{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, err
	}

	// If the user is active then we can return their media lists
	if user.IsActive {
		query := datastore.NewQuery("MediaList").Filter("CreatedBy =", user.Id).Filter("Archived =", archived)
		query = constructQuery(query, r)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.MediaList{}, nil, 0, err
		}

		mediaLists = make([]models.MediaList, len(ks))
		err = nds.GetMulti(c, ks, mediaLists)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.MediaList{}, nil, 0, err
		}

		for i := 0; i < len(mediaLists); i++ {
			mediaLists[i].Format(ks[i], "lists")
			mediaLists[i].AddNewFieldsMapToOldLists(c)
		}
	}

	// If the user is not active then we block their media lists
	return mediaLists, nil, len(mediaLists), nil
}

// Gets every single media list
func GetPublicMediaLists(c context.Context, r *http.Request) ([]models.MediaList, interface{}, int, error) {
	mediaLists := []models.MediaList{}

	_, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, err
	}

	query := datastore.NewQuery("MediaList").Filter("PublicList =", true).Filter("Archived =", false)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, err
	}

	mediaLists = make([]models.MediaList, len(ks))
	err = nds.GetMulti(c, ks, mediaLists)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, nil, 0, err
	}

	for i := 0; i < len(mediaLists); i++ {
		mediaLists[i].Format(ks[i], "lists")

		if mediaLists[i].PublicList {
			mediaLists[i].ReadOnly = true
		}
	}

	return mediaLists, nil, len(mediaLists), nil
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

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return medialist, nil, err
	}

	// Initial values for fieldsmap
	medialist.FieldsMap = getFieldsMap()

	// Create media list
	_, err = medialist.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	return medialist, nil, nil
}

func CreateSampleMediaList(c context.Context, r *http.Request, user models.User) (models.MediaList, interface{}, error) {
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
	_, err := Create(c, r, &singleContact)
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
	_, err = Create(c, r, &fashionContact)
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
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}
	if mediaList.CreatedBy != user.Id {
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

	utilities.UpdateIfNotBlank(&mediaList.Name, updatedMediaList.Name)

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
	sync.ResourceSync(r, mediaList.Id, "List", "update")
	return mediaList, nil, nil
}

/*
* Action methods
 */

func GetContactsForList(c context.Context, r *http.Request, id string) ([]models.Contact, interface{}, int, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	queryField := gcontext.Get(r, "q").(string)
	if queryField != "" {
		contacts, err := search.SearchContactsByList(c, r, queryField, user, mediaList.Id)
		if err != nil {
			return []models.Contact{}, nil, 0, err
		}

		publications := contactsToPublications(c, contacts)
		return contacts, publications, len(contacts), nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	startPosition := offset
	endPosition := startPosition + limit

	if len(mediaList.Contacts) < startPosition {
		return []models.Contact{}, nil, 0, err
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
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	instagramUsers := []string{}
	twitterUsers := []string{}

	for i := 0; i < len(contacts); i++ {
		contacts[i].Id = subsetIds[i]
		contacts[i].Type = "contacts"

		if contacts[i].Instagram != "" {
			instagramUsers = append(instagramUsers, contacts[i].Instagram)
		}

		if contacts[i].Twitter != "" {
			twitterUsers = append(twitterUsers, contacts[i].Twitter)
		}
	}

	instagramTimeseries := []search.InstagramTimeseries{}

	// Check if there are special fields we need to get data for
	for i := 0; i < len(mediaList.FieldsMap); i++ {
		if !mediaList.FieldsMap[i].Hidden && mediaList.FieldsMap[i].CustomField {
			if mediaList.FieldsMap[i].Name == "instagramfollowers" {
				instagramTimeseries, _ = search.SearchInstagramTimeseriesByUsernames(c, r, instagramUsers)
			}
		}
	}

	if len(instagramTimeseries) > 0 {
		customFieldNameToValue := map[string]search.InstagramTimeseries{}
		for i := 0; i < len(instagramTimeseries); i++ {
			lowerCaseUsername := strings.ToLower(instagramTimeseries[i].Username)
			customFieldNameToValue[lowerCaseUsername] = instagramTimeseries[i]
		}

		for i := 0; i < len(contacts); i++ {
			customField := models.CustomContactField{}
			customField.Name = "instagramfollowers"
			customField.Value = strconv.Itoa(customFieldNameToValue[contacts[i].Instagram].Followers)
			contacts[i].CustomFields = append(contacts[i].CustomFields, customField)
		}
	}

	log.Infof(c, "%v", instagramTimeseries)

	// Add includes
	publications := contactsToPublications(c, contacts)
	return contacts, publications, len(contacts), nil
}

func GetEmailsForList(c context.Context, r *http.Request, id string) ([]models.Email, interface{}, int, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	emails, count, err := filterEmailbyListId(c, r, mediaList.Id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, nil, 0, err
	}

	return emails, nil, count, nil
}

func GetHeadlinesForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ListId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	headlines, err := search.SearchHeadlinesByResourceId(c, r, feeds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return headlines, nil, len(headlines), nil
}

func GetFeedForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
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
		return []models.Contact{}, nil, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ListId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	feed, err := search.SearchFeedForContacts(c, r, contacts, feeds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return feed, nil, len(feed), nil
}

func GetTweetsForList(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
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
		return []models.Contact{}, nil, 0, err
	}

	usernames := []string{}
	for i := 0; i < len(contacts); i++ {
		if contacts[i].Twitter != "" {
			usernames = append(usernames, contacts[i].Twitter)
		}
	}

	tweets, err := search.SearchTweetsByUsernames(c, r, usernames)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return tweets, nil, len(tweets), nil
}

func DuplicateList(c context.Context, r *http.Request, id string) (models.MediaList, interface{}, error) {
	// Get the details of the current media list
	mediaList, _, err := GetMediaList(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.MediaList{}, nil, err
	}
	if mediaList.CreatedBy != user.Id {
		return models.MediaList{}, nil, errors.New("Forbidden")
	}

	return models.MediaList{}, nil, nil
}
