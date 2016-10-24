package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getContact(c context.Context, r *http.Request, id int64) (models.Contact, error) {
	if id == 0 {
		return models.Contact{}, errors.New("datastore: no such entity")
	}
	// Get the Contact by id
	var contact models.Contact
	contactId := datastore.NewKey(c, "Contact", "", id, nil)
	err := nds.Get(c, contactId, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	if !contact.Created.IsZero() {
		contact.Format(contactId, "contacts")

		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		if !contact.IsMasterContact && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		return contact, nil
	}
	return models.Contact{}, errors.New("No contact by this id")
}

/*
* Update methods
 */

func updateContact(c context.Context, r *http.Request, contact *models.Contact, updatedContact models.Contact) (models.Contact, interface{}, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return *contact, nil, err
	}

	// Check if the old Twitter is changed to a new one
	// If both of them are not empty but also not the same
	if contact.Twitter != "" && updatedContact.Twitter != "" && contact.Twitter != updatedContact.Twitter {
		updatedContact.Normalize()
		contact.TwitterPrivate = false
		contact.TwitterInvalid = false
		sync.TwitterSync(r, updatedContact.Twitter)
	}

	// If you are changing Instagram usernames
	if contact.Instagram != "" && updatedContact.Instagram != "" && contact.Instagram != updatedContact.Instagram {
		contact.InstagramPrivate = false
		contact.InstagramInvalid = false
		sync.InstagramSync(r, updatedContact.Instagram, currentUser.InstagramAuthKey)
	}

	if contact.Twitter == "" && updatedContact.Twitter != "" {
		updatedContact.Normalize()
		contact.TwitterPrivate = false
		contact.TwitterInvalid = false
		sync.TwitterSync(r, updatedContact.Twitter)
	}

	// If they add a new Instagram
	if contact.Instagram == "" && updatedContact.Instagram != "" {
		updatedContact.Normalize()
		contact.InstagramPrivate = false
		contact.InstagramInvalid = false
		sync.InstagramSync(r, updatedContact.Instagram, currentUser.InstagramAuthKey)
	}

	utilities.UpdateIfNotBlank(&contact.FirstName, updatedContact.FirstName)
	utilities.UpdateIfNotBlank(&contact.LastName, updatedContact.LastName)
	utilities.UpdateIfNotBlank(&contact.Email, updatedContact.Email)
	utilities.UpdateIfNotBlank(&contact.LinkedIn, updatedContact.LinkedIn)
	utilities.UpdateIfNotBlank(&contact.Twitter, updatedContact.Twitter)
	utilities.UpdateIfNotBlank(&contact.Instagram, updatedContact.Instagram)
	utilities.UpdateIfNotBlank(&contact.Website, updatedContact.Website)
	utilities.UpdateIfNotBlank(&contact.Blog, updatedContact.Blog)
	utilities.UpdateIfNotBlank(&contact.Notes, updatedContact.Notes)
	utilities.UpdateIfNotBlank(&contact.Location, updatedContact.Location)
	utilities.UpdateIfNotBlank(&contact.PhoneNumber, updatedContact.PhoneNumber)

	if updatedContact.ListId != 0 {
		contact.ListId = updatedContact.ListId
	}

	if len(updatedContact.CustomFields) > 0 {
		contact.CustomFields = updatedContact.CustomFields
	}

	if len(updatedContact.Employers) > 0 {
		contact.Employers = updatedContact.Employers
	}

	if len(updatedContact.PastEmployers) > 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	// Special case when you want to remove all the employers
	if len(contact.Employers) > 0 && len(updatedContact.Employers) == 0 {
		contact.Employers = updatedContact.Employers
	}

	// Special case when you want to remove all the past employers
	if len(contact.PastEmployers) > 0 && len(updatedContact.PastEmployers) == 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	_, err = Save(c, r, contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	return *contact, nil, nil
}

/*
* Filter methods
 */

func filterContacts(c context.Context, r *http.Request, queryType, query string) ([]models.Contact, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, err
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, err
	}

	if len(contacts) > 0 {
		for i := 0; i < len(contacts); i++ {
			contacts[i].Format(ks[i], "contacts")
		}

		return contacts, nil

	}
	return []models.Contact{}, errors.New("No contact by this " + queryType)
}

func filterContact(c context.Context, r *http.Request, queryType, query string) (models.Contact, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("Contact").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, err
	}

	if len(contacts) > 0 {
		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		if !contacts[0].IsMasterContact && !permissions.AccessToObject(contacts[0].CreatedBy, user.Id) {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}
		contacts[0].Format(ks[0], "contacts")
		return contacts[0], nil
	}
	return models.Contact{}, errors.New("No contact by this " + queryType)
}

/*
* Normalization methods
 */

func filterContactByEmail(c context.Context, email string) ([]models.Contact, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("Contact").Filter("Email =", email).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, err
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, err
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].Format(ks[i], "contacts")
	}

	return contacts, nil
}

func contactsToPublications(c context.Context, contacts []models.Contact) []models.Publication {
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

func contactsToLists(c context.Context, r *http.Request, contacts []models.Contact) []models.MediaList {
	mediaListIds := []int64{}

	for i := 0; i < len(contacts); i++ {
		mediaListIds = append(mediaListIds, contacts[i].ListId)
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

func getIncludesForContact(c context.Context, r *http.Request, contacts []models.Contact) interface{} {
	mediaLists := contactsToLists(c, r, contacts)
	publications := contactsToPublications(c, contacts)

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
* Public methods
 */

/*
* Get methods
 */

// Gets every single contact
func GetContacts(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	// If the user is currently active
	if user.IsActive {
		queryField := gcontext.Get(r, "q").(string)
		if queryField != "" {
			contacts, err := search.SearchContacts(c, r, queryField, user.Id)
			if err != nil {
				return []models.Contact{}, nil, 0, err
			}
			includes := getIncludesForContact(c, r, contacts)
			return contacts, includes, len(contacts), nil
		}

		query := datastore.NewQuery("Contact").Filter("CreatedBy =", user.Id).Filter("IsMasterContact = ", false)
		query = constructQuery(query, r)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		contacts := []models.Contact{}
		contacts = make([]models.Contact, len(ks))
		err = nds.GetMulti(c, ks, contacts)
		if err != nil {
			log.Errorf(c, "%v", err)
			return contacts, nil, 0, err
		}

		for i := 0; i < len(contacts); i++ {
			contacts[i].Format(ks[i], "contacts")
		}

		includes := getIncludesForContact(c, r, contacts)
		return contacts, includes, len(contacts), nil
	}

	// If user is not active then return empty lists
	return []models.Contact{}, nil, 0, nil
}

func GetContact(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	includes := getIncludesForContact(c, r, []models.Contact{contact})
	return contact, includes, nil
}

func GetTweetsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	tweets, err := search.SearchTweetsByUsername(c, r, contact.Twitter)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return tweets, nil, len(tweets), nil
}

func GetTwitterProfileForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	twitterProfile, err := search.SearchProfileByUsername(c, r, contact.Twitter)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return twitterProfile, nil, nil
}

func GetInstagramPostsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	instagramPosts, err := search.SearchInstagramPostsByUsername(c, r, contact.Instagram)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return instagramPosts, nil, len(instagramPosts), nil
}

func GetInstagramProfileForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	instagramProfile, err := search.SearchInstagramProfileByUsername(c, r, contact.Instagram)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return instagramProfile, nil, nil
}

func GetInstagramTimeseriesForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	instagramTimeseries, err := search.SearchInstagramTimeseriesByUsername(c, r, contact.Instagram)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return instagramTimeseries, nil, nil
}

func GetTwitterTimeseriesForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	twitterTimeseries, err := search.SearchTwitterTimeseriesByUsername(c, r, contact.Twitter)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return twitterTimeseries, nil, nil
}

func GetEmailsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	// To check if the user can access it
	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	emails, err := filterEmailbyContactId(c, r, contact.Id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return emails, nil, len(emails), nil
}

func GetHeadlinesForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ContactId", currentId)
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

func GetFeedForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ContactId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	feed, err := search.SearchFeedForContacts(c, r, []models.Contact{contact}, feeds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return feed, nil, len(feed), nil
}

func GetFeedsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ContactId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return feeds, nil, len(feeds), nil
}

func GetSimilarContacts(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	allKeysMap := map[*datastore.Key]bool{}

	if contact.LinkedIn != "" {
		query := datastore.NewQuery("Contact").Filter("LinkedIn =", contact.LinkedIn).Filter("CreatedBy = ", currentUser.Id).Filter("IsMasterContact =", false)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, 0, err
		}

		for i := 0; i < len(ks); i++ {
			allKeysMap[ks[i]] = true
		}
	}

	if contact.Twitter != "" {
		query := datastore.NewQuery("Contact").Filter("Twitter =", contact.Twitter).Filter("CreatedBy = ", currentUser.Id).Filter("IsMasterContact =", false)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, 0, err
		}

		for i := 0; i < len(ks); i++ {
			allKeysMap[ks[i]] = true
		}
	}

	if contact.Instagram != "" {
		query := datastore.NewQuery("Contact").Filter("Instagram =", contact.Instagram).Filter("CreatedBy = ", currentUser.Id).Filter("IsMasterContact =", false)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, 0, err
		}

		for i := 0; i < len(ks); i++ {
			allKeysMap[ks[i]] = true
		}
	}

	if contact.Website != "" {
		query := datastore.NewQuery("Contact").Filter("Website =", contact.Website).Filter("CreatedBy = ", currentUser.Id).Filter("IsMasterContact =", false)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, 0, err
		}

		for i := 0; i < len(ks); i++ {
			allKeysMap[ks[i]] = true
		}
	}

	if contact.Blog != "" {
		query := datastore.NewQuery("Contact").Filter("Blog =", contact.Blog).Filter("CreatedBy = ", currentUser.Id).Filter("IsMasterContact =", false)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return nil, nil, 0, err
		}

		for i := 0; i < len(ks); i++ {
			allKeysMap[ks[i]] = true
		}
	}

	allKeys := []*datastore.Key{}
	for k := range allKeysMap {
		allKeys = append(allKeys, k)
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	startPosition := offset
	endPosition := startPosition + limit

	if len(allKeys) < startPosition {
		return []models.Contact{}, nil, 0, err
	}

	if len(allKeys) < endPosition {
		endPosition = len(allKeys)
	}

	subsetIds := allKeys[startPosition:endPosition]
	contacts := []models.Contact{}
	contacts = make([]models.Contact, len(subsetIds))
	err = nds.GetMulti(c, subsetIds, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contacts, nil, 0, err
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].Format(subsetIds[i], "contacts")
	}

	return contacts, nil, len(contacts), nil
}

func FilterContacts(c context.Context, r *http.Request, queryType, query string) ([]models.Contact, error) {
	// User has to be logged in
	_, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Contact{}, err
	}

	return filterContacts(c, r, queryType, query)
}

/*
* Create methods
 */

func Create(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return ct, err
	}

	ct.Create(c, r, currentUser)
	_, err = Save(c, r, ct)

	// Sync with ES
	sync.ResourceSync(r, ct.Id, "Contact", "create")

	// If user is just created
	if ct.Twitter != "" {
		sync.TwitterSync(r, ct.Twitter)
	}
	if ct.Instagram != "" {
		sync.InstagramSync(r, ct.Instagram, currentUser.InstagramAuthKey)
	}

	return ct, err
}

func CreateContact(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	decoder := ffjson.NewDecoder()
	var contact models.Contact
	err := decoder.Decode(buf, &contact)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var contacts []models.Contact

		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &contacts)

		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		newContacts := []models.Contact{}
		for i := 0; i < len(contacts); i++ {
			_, err = Create(c, r, &contacts[i])
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Contact{}, nil, 0, err
			}
			newContacts = append(newContacts, contacts[i])
		}

		return newContacts, nil, len(newContacts), nil
	}

	// Create contact
	_, err = Create(c, r, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	return []models.Contact{contact}, nil, 0, nil
}

// Does a ES sync in parse package & Twitter sync here
func BatchCreateContactsForExcelUpload(c context.Context, r *http.Request, contacts []models.Contact, mediaListId int64) ([]int64, error) {
	var keys []*datastore.Key
	var contactIds []int64

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, err
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].CreatedBy = currentUser.Id
		contacts[i].Created = time.Now()
		contacts[i].Updated = time.Now()
		contacts[i].ListId = mediaListId
		contacts[i].Normalize()
		keys = append(keys, contacts[i].Key(c))

		// If there is a Twitter
		if contacts[i].Twitter != "" {
			sync.TwitterSync(r, contacts[i].Twitter)
		}
		if contacts[i].Instagram != "" {
			sync.InstagramSync(r, contacts[i].Instagram, currentUser.InstagramAuthKey)
		}
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

	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, err
	}

	for i := 0; i < len(ks); i++ {
		contactIds = append(contactIds, ks[i].IntID())
	}

	return contactIds, nil
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func Save(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	// Update the Updated time
	ct.Normalize()
	ct.Save(c, r)
	sync.ResourceSync(r, ct.Id, "Contact", "create")
	return ct, nil
}

func UpdateSingleContact(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get the details of the current contact
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, errors.New("Could not get user")
	}

	if !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
		return models.Contact{}, nil, errors.New("You don't have permissions to edit these objects")
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedContact models.Contact
	err = decoder.Decode(buf, &updatedContact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	return updateContact(c, r, &contact, updatedContact)
}

func UpdateBatchContact(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedContacts []models.Contact
	err := decoder.Decode(buf, &updatedContacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, err
	}

	// Get logged in user
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, errors.New("Could not get user")
	}

	// Check if each of the contacts have permissions before updating anything
	currentContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		contact, err := getContact(c, r, updatedContacts[i].Id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		if !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			return []models.Contact{}, nil, 0, errors.New("Forbidden")
		}

		currentContacts = append(currentContacts, contact)
	}

	// Update each of the contacts
	newContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		updatedContact, _, err := updateContact(c, r, &currentContacts[i], updatedContacts[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, err
		}

		newContacts = append(newContacts, updatedContact)
	}

	return newContacts, nil, len(newContacts), nil
}

/*
* Delete methods
 */

func DeleteContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	// Double check permissions. Admins should not be able to delete.
	if !permissions.AccessToObject(contact.CreatedBy, user.Id) {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	contact.IsDeleted = true
	contact.Save(c, r)

	// Pubsub to remove ES contact
	sync.ResourceSync(r, contact.Id, "Contact", "delete")

	return nil, nil, nil
}
