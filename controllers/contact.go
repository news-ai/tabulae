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

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"

	"github.com/news-ai/api/controllers"
	apiSearch "github.com/news-ai/api/search"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"
)

/*
* Private methods
 */

type copyContactsDetails struct {
	Contacts []int64 `json:"contacts"`
	ListId   int64   `json:"listid"`
}

type deleteContactsDetails struct {
	Contacts []int64 `json:"contacts"`
}

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

		user, err := controllers.GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		contactList, err := getMediaList(c, r, contact.ListId)
		if err != nil {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		// If it is a public list
		if contact.ListId != 0 && contactList.PublicList {
			// You don't own it and you are not an admin
			if contactList.PublicList && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
				contact.ReadOnly = true
			}
			return contact, nil
		}

		// This runs if it is not a public list
		if contactList.TeamId != user.TeamId && !contact.IsMasterContact && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		return contact, nil
	}
	return models.Contact{}, errors.New("No contact by this id")
}

/*
* Create methods
 */

func createContact(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return ct, err
	}

	ct.FormatName()
	ct.Normalize()

	_, err = enrichContact(c, r, ct)
	if err != nil {
		log.Errorf(c, "%v", err)
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

/*
* Update methods
 */

func updateContact(c context.Context, r *http.Request, contact *models.Contact, updatedContact models.Contact) (models.Contact, interface{}, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
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

	previousEmail := contact.Email

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

	if contact.Email != previousEmail {
		_, err = enrichContact(c, r, contact)
		if err != nil {
			log.Errorf(c, "%v", err)
		}
	}

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

	if len(updatedContact.Tags) > 0 {
		contact.Tags = updatedContact.Tags
	}

	// Special case when you want to remove all the employers
	if len(contact.Employers) > 0 && len(updatedContact.Employers) == 0 {
		contact.Employers = updatedContact.Employers
	}

	// Special case when you want to remove all the past employers
	if len(contact.PastEmployers) > 0 && len(updatedContact.PastEmployers) == 0 {
		contact.PastEmployers = updatedContact.PastEmployers
	}

	// Special case when you want to remove all the past employers
	if len(contact.Tags) > 0 && len(updatedContact.Tags) == 0 {
		contact.Tags = updatedContact.Tags
	}

	_, err = Save(c, r, contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	// When editing a contact on the list view we need the timeseries data in it
	if contact.ListId == 0 {
		return *contact, nil, nil
	}

	mediaList, err := getMediaList(c, r, contact.ListId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return *contact, nil, nil
	}

	contacts, err := ContactsToDefaultFields(c, r, []models.Contact{*contact}, mediaList)

	return contacts[0], nil, nil
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
		user, err := controllers.GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		mediaList, err := getMediaList(c, r, contacts[0].ListId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}

		if !contacts[0].IsMasterContact && mediaList.TeamId != user.TeamId && !permissions.AccessToObject(contacts[0].CreatedBy, user.Id) {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.Contact{}, err
		}
		contacts[0].Format(ks[0], "contacts")
		return contacts[0], nil
	}
	return models.Contact{}, errors.New("No contact by this " + queryType)
}

func filterListsbyContactEmail(c context.Context, r *http.Request, email string) ([]models.MediaList, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, err
	}

	ks, err := datastore.NewQuery("Contact").Filter("CreatedBy =", user.Id).Filter("Email =", email).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, err
	}

	var contacts []models.Contact
	contacts = make([]models.Contact, len(ks))
	err = nds.GetMulti(c, ks, contacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.MediaList{}, err
	}

	mediaListsIds := []int64{}
	mediaLists := []models.MediaList{}
	if len(contacts) > 0 {
		for i := 0; i < len(contacts); i++ {
			if contacts[i].ListId != 0 && !contacts[i].IsDeleted {
				mediaListsIds = append(mediaListsIds, contacts[i].ListId)
			}
		}

		mediaListAdded := map[int64]bool{}
		for i := 0; i < len(mediaListsIds); i++ {
			if _, ok := mediaListAdded[mediaListsIds[i]]; !ok {
				singleMediaList, err := getMediaList(c, r, mediaListsIds[i])
				if err == nil && !singleMediaList.Archived {
					mediaLists = append(mediaLists, singleMediaList)
					mediaListAdded[mediaListsIds[i]] = true
				}
			}
		}

		return mediaLists, nil
	}

	return []models.MediaList{}, errors.New("No media lists for this email")
}

func filterContactsForListId(c context.Context, r *http.Request, listId int64) ([]models.Contact, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("Contact").Filter("ListId =", listId).Filter("IsDeleted =", false).KeysOnly().GetAll(c, nil)
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

	return []models.Contact{}, errors.New("No contact by this ListId")
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

func filterContactByEmailForUser(c context.Context, r *http.Request, id int64) ([]models.Contact, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, err
	}

	contact, err := getContact(c, r, id)
	if err != nil {
		return []models.Contact{}, err
	}

	ks, err := datastore.NewQuery("Contact").Filter("CreatedBy =", user.Id).Filter("Email =", contact.Email).KeysOnly().GetAll(c, nil)
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

	// return everything but the current contact
	fitleredContacts := []models.Contact{}
	for i := 0; i < len(contacts); i++ {
		contacts[i].Format(ks[i], "contacts")
		if contacts[i].Id != contact.Id {
			fitleredContacts = append(fitleredContacts, contacts[i])
		}
	}

	return fitleredContacts, nil
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

func getIncludesForContacts(c context.Context, r *http.Request, contacts []models.Contact) interface{} {
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
* Action methods
 */

func enrichContact(c context.Context, r *http.Request, contact *models.Contact) (interface{}, error) {
	if contact.Email == "" {
		return nil, errors.New("Contact does not have an email")
	}

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	contactDetail, err := apiSearch.SearchContactDatabase(c, r, contact.Email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	if contactDetail.Data.Likelihood > 0.75 {
		// Add first name
		if contact.FirstName == "" && contactDetail.Data.ContactInfo.GivenName != "" {
			contact.FirstName = contactDetail.Data.ContactInfo.GivenName
		}

		// Add last name
		if contact.LastName == "" && contactDetail.Data.ContactInfo.FamilyName != "" {
			contact.LastName = contactDetail.Data.ContactInfo.FamilyName
		}

		// Add social profiles
		if len(contactDetail.Data.SocialProfiles) > 0 {
			for i := 0; i < len(contactDetail.Data.SocialProfiles); i++ {
				if contactDetail.Data.SocialProfiles[i].TypeID == "linkedin" {
					if contact.LinkedIn == "" {
						contact.LinkedIn = contactDetail.Data.SocialProfiles[i].URL
					}
				}

				if contactDetail.Data.SocialProfiles[i].TypeID == "twitter" {
					if contact.Twitter == "" {
						contact.Twitter = contactDetail.Data.SocialProfiles[i].Username
						sync.TwitterSync(r, contact.Twitter)
					}
				}

				if contactDetail.Data.SocialProfiles[i].TypeID == "instagram" {
					if contact.Instagram == "" {
						contact.Instagram = contactDetail.Data.SocialProfiles[i].URL
						sync.InstagramSync(r, contact.Instagram, currentUser.InstagramAuthKey)
					}
				}
			}
		}

		// Add jobs
		if len(contactDetail.Data.Organizations) > 0 {
			for i := 0; i < len(contactDetail.Data.Organizations); i++ {
				if contactDetail.Data.Organizations[i].Name != "" {
					// Get which publication it is in our database
					publication, err := FindOrCreatePublication(c, r, contactDetail.Data.Organizations[i].Name, "")
					if err != nil {
						log.Errorf(c, "%v", err)
						continue
					}

					previousJob := false

					// Check if this position was in the past or current
					if contactDetail.Data.Organizations[i].EndDate != "" {
						previousJob = true
					}

					alreadyExists := false
					for x := 0; x < len(contact.Employers); x++ {
						currentPublication, err := getPublication(c, contact.Employers[x])
						if err != nil {
							log.Errorf(c, "%v", err)
							continue
						}

						if currentPublication.Name == publication.Name {
							alreadyExists = true
						}

						if currentPublication.Id == publication.Id {
							alreadyExists = true
						}
					}

					for x := 0; x < len(contact.PastEmployers); x++ {
						currentPublication, err := getPublication(c, contact.PastEmployers[x])
						if err != nil {
							log.Errorf(c, "%v", err)
							continue
						}

						if currentPublication.Name == publication.Name {
							alreadyExists = true
						}

						if currentPublication.Id == publication.Id {
							alreadyExists = true
						}
					}

					// Add to list
					if !alreadyExists {
						if previousJob {
							contact.PastEmployers = append(contact.PastEmployers, publication.Id)
						} else {
							contact.Employers = append(contact.Employers, publication.Id)
						}
					}

				}
			}
		}

		// Add location
		contact.Location = contactDetail.Data.Demographics.LocationDeduced.NormalizedLocation

		// Add website
		if len(contactDetail.Data.ContactInfo.Websites) > 0 {
			contact.Website = contactDetail.Data.ContactInfo.Websites[0].URL
		}

		// Add tags
		if len(contactDetail.Data.DigitalFootprint.Topics) > 0 {
			tags := []string{}
			for i := 0; i < len(contactDetail.Data.DigitalFootprint.Topics); i++ {
				if contactDetail.Data.DigitalFootprint.Topics[i].Value != "" {
					tags = append(tags, contactDetail.Data.DigitalFootprint.Topics[i].Value)
				}
			}

			contact.Tags = append(contact.Tags, tags...)

			// Remove duplicates from tags
			found := make(map[string]bool)
			j := 0
			for i, x := range contact.Tags {
				if !found[x] {
					found[x] = true
					contact.Tags[j] = contact.Tags[i]
					j++
				}
			}
			contact.Tags = contact.Tags[:j]
		}

		return nil, nil
	}

	return nil, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single contact
func GetContacts(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	// If the user is currently active
	if user.IsActive {
		queryField := gcontext.Get(r, "q").(string)
		if queryField != "" {
			fieldSelector := strings.Split(queryField, ":")
			if len(fieldSelector) != 2 {
				contacts, total, err := search.SearchContacts(c, r, queryField, user.Id)
				if err != nil {
					return []models.Contact{}, nil, 0, 0, err
				}
				includes := getIncludesForContacts(c, r, contacts)
				return contacts, includes, len(contacts), total, nil
			} else {
				selectedContacts, total, err := search.SearchContactsByFieldSelector(c, r, fieldSelector[0], fieldSelector[1], user.Id)
				if err != nil {
					return nil, nil, 0, 0, err
				}

				contacts := []models.Contact{}
				for i := 0; i < len(selectedContacts); i++ {
					singleContact, err := getContact(c, r, selectedContacts[i].Id)
					if err == nil {
						contacts = append(contacts, singleContact)
					}
				}

				includes := getIncludesForContacts(c, r, contacts)
				return contacts, includes, len(contacts), total, nil
			}
		}

		query := datastore.NewQuery("Contact").Filter("CreatedBy =", user.Id).Filter("IsMasterContact = ", false)
		query = controllers.ConstructQuery(query, r)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, 0, err
		}

		contacts := []models.Contact{}
		contacts = make([]models.Contact, len(ks))
		err = nds.GetMulti(c, ks, contacts)
		if err != nil {
			log.Errorf(c, "%v", err)
			return contacts, nil, 0, 0, err
		}

		for i := 0; i < len(contacts); i++ {
			contacts[i].Format(ks[i], "contacts")
		}

		includes := getIncludesForContacts(c, r, contacts)
		return contacts, includes, len(contacts), 0, nil
	}

	// If user is not active then return empty lists
	return []models.Contact{}, nil, 0, 0, nil
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

	includes := getIncludesForContacts(c, r, []models.Contact{contact})
	return contact, includes, nil
}

func ContactsToDefaultFields(c context.Context, r *http.Request, contacts []models.Contact, mediaList models.MediaList) ([]models.Contact, error) {
	instagramUsers := []string{}
	twitterUsers := []string{}

	for i := 0; i < len(contacts); i++ {
		if contacts[i].Instagram != "" {
			instagramUsers = append(instagramUsers, contacts[i].Instagram)
		}

		if contacts[i].Twitter != "" {
			twitterUsers = append(twitterUsers, contacts[i].Twitter)
		}
	}

	readOnlyPresent := []string{}
	instagramTimeseries := []apiSearch.InstagramTimeseries{}
	twitterTimeseries := []apiSearch.TwitterTimeseries{}

	// Check if there are special fields we need to get data for
	for i := 0; i < len(mediaList.FieldsMap); i++ {
		if mediaList.FieldsMap[i].ReadOnly && !mediaList.FieldsMap[i].Hidden {
			readOnlyPresent = append(readOnlyPresent, mediaList.FieldsMap[i].Value)
			if strings.Contains(mediaList.FieldsMap[i].Value, "instagram") {
				if len(instagramTimeseries) == 0 {
					instagramTimeseries, _ = apiSearch.SearchInstagramTimeseriesByUsernames(c, r, instagramUsers)
				}
			}
			if strings.Contains(mediaList.FieldsMap[i].Value, "twitter") {
				if len(twitterTimeseries) == 0 {
					twitterTimeseries, _ = apiSearch.SearchTwitterTimeseriesByUsernames(c, r, twitterUsers)
				}
			}
		}
	}

	if len(readOnlyPresent) > 0 {
		customFieldInstagramUsernameToValue := map[string]apiSearch.InstagramTimeseries{}
		customFieldTwitterUsernameToValue := map[string]apiSearch.TwitterTimeseries{}

		if len(instagramTimeseries) > 0 {
			for i := 0; i < len(instagramTimeseries); i++ {
				lowerCaseUsername := strings.ToLower(instagramTimeseries[i].Username)
				customFieldInstagramUsernameToValue[lowerCaseUsername] = instagramTimeseries[i]
			}
		}

		if len(twitterTimeseries) > 0 {
			for i := 0; i < len(twitterTimeseries); i++ {
				lowerCaseUsername := strings.ToLower(twitterTimeseries[i].Username)
				customFieldTwitterUsernameToValue[lowerCaseUsername] = twitterTimeseries[i]
			}
		}

		for i := 0; i < len(contacts); i++ {
			for x := 0; x < len(readOnlyPresent); x++ {
				customField := models.CustomContactField{}
				customField.Name = readOnlyPresent[x]

				lowerCaseInstagramUsername := strings.ToLower(contacts[i].Instagram)
				lowerCaseTwitterUsername := strings.ToLower(contacts[i].Twitter)

				if lowerCaseInstagramUsername != "" {
					if _, ok := customFieldInstagramUsernameToValue[lowerCaseInstagramUsername]; ok {
						instagramProfile := customFieldInstagramUsernameToValue[lowerCaseInstagramUsername]

						if customField.Name == "instagramfollowers" {
							customField.Value = strconv.Itoa(instagramProfile.Followers)
						} else if customField.Name == "instagramfollowing" {
							customField.Value = strconv.Itoa(instagramProfile.Following)
						} else if customField.Name == "instagramlikes" {
							customField.Value = strconv.Itoa(instagramProfile.Likes)
						} else if customField.Name == "instagramcomments" {
							customField.Value = strconv.Itoa(instagramProfile.Comments)
						} else if customField.Name == "instagramposts" {
							customField.Value = strconv.Itoa(instagramProfile.Posts)
						}
					}
				}

				if lowerCaseTwitterUsername != "" {
					if _, ok := customFieldTwitterUsernameToValue[lowerCaseTwitterUsername]; ok {
						twitterProfile := customFieldTwitterUsernameToValue[lowerCaseTwitterUsername]

						if customField.Name == "twitterfollowers" {
							customField.Value = strconv.Itoa(twitterProfile.Followers)
						} else if customField.Name == "twitterfollowing" {
							customField.Value = strconv.Itoa(twitterProfile.Following)
						} else if customField.Name == "twitterlikes" {
							customField.Value = strconv.Itoa(twitterProfile.Likes)
						} else if customField.Name == "twitterretweets" {
							customField.Value = strconv.Itoa(twitterProfile.Retweets)
						} else if customField.Name == "twitterposts" {
							customField.Value = strconv.Itoa(twitterProfile.Posts)
						}
					}
				}

				if customField.Name == "latestheadline" {
					// Get the feed of the contact
					headlines, _, _, _, err := GetHeadlinesForContactById(c, r, contacts[i].Id)

					// Set the value of the post name to the user
					if err == nil && len(headlines) > 0 {
						customField.Value = headlines[0].Title
					}
				}

				if customField.Name == "lastcontacted" {
					emails, _, _, err := GetOrderedEmailsForContact(c, r, contacts[i])

					// Set the value of the post name to the user
					if err == nil && len(emails) > 0 {
						lastUnarchivedEmail := -1
						for y := 0; y < len(emails); y++ {
							if !emails[y].Archived && !emails[y].Cancel {
								lastUnarchivedEmail = y
								break
							}
						}
						// The processing here is a little more complex
						// customField.Value = emails[0].Created
						// Also, don't count it if the email has been
						// archived. So we look for the last unarchived email.
						if lastUnarchivedEmail != -1 && lastUnarchivedEmail < len(emails) {
							if !emails[lastUnarchivedEmail].SendAt.IsZero() {
								customField.Value = emails[lastUnarchivedEmail].SendAt.Format(time.RFC3339)
							} else {
								customField.Value = emails[lastUnarchivedEmail].Created.Format(time.RFC3339)
							}

							// Outliar checks
							// First check if email is sent & delivered.
							if customField.Value != "" && emails[lastUnarchivedEmail].IsSent && emails[lastUnarchivedEmail].Delievered {
								// Sometimes email is marked as sent, but hasn't actually been sent
								// because Gmail rejected it. This is to check that.
								if emails[lastUnarchivedEmail].Method == "gmail" && emails[lastUnarchivedEmail].GmailId == "" {
									customField.Value = ""
								}
							}
						}
					}
				}

				if customField.Value != "" {
					contacts[i].CustomFields = append(contacts[i].CustomFields, customField)
				}
			}
		}
	}

	return contacts, nil
}

func GetTweetsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	tweets, total, err := apiSearch.SearchTweetsByUsername(c, r, contact.Twitter)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return tweets, nil, len(tweets), total, nil
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

	twitterProfile, err := apiSearch.SearchProfileByUsername(c, r, contact.Twitter)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return twitterProfile, nil, nil
}

func GetInstagramPostsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	instagramPosts, total, err := apiSearch.SearchInstagramPostsByUsername(c, r, contact.Instagram)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return instagramPosts, nil, len(instagramPosts), total, nil
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

	instagramProfile, err := apiSearch.SearchInstagramProfileByUsername(c, r, contact.Instagram)
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

	instagramTimeseries, _, err := apiSearch.SearchInstagramTimeseriesByUsername(c, r, contact.Instagram)
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

	twitterTimeseries, _, err := apiSearch.SearchTwitterTimeseriesByUsername(c, r, contact.Twitter)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return twitterTimeseries, nil, nil
}

func GetOrderedEmailsForContact(c context.Context, r *http.Request, contact models.Contact) ([]models.Email, interface{}, int, error) {
	// To check if the user can access it
	emails, err := filterOrderedEmailbyContactId(c, r, contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return emails, nil, len(emails), nil
}

func GetEmailsForContactById(c context.Context, r *http.Request, currentId int64) ([]models.Email, interface{}, int, int, error) {
	// To check if the user can access it
	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	if contact.Email == "" {
		return []models.Email{}, nil, 0, 0, nil
	}

	emails, err := filterEmailbyContactEmail(c, r, contact.Email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return emails, nil, len(emails), 0, nil
}

func GetEmailsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return GetEmailsForContactById(c, r, currentId)
}

func GetListsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	// To check if the user can access it
	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	if contact.Email == "" {
		return []models.MediaList{}, nil, 0, 0, errors.New("Contact has no email")
	}

	mediaLists, err := filterListsbyContactEmail(c, r, contact.Email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return mediaLists, nil, len(mediaLists), 0, nil
}

func GetHeadlinesForContactById(c context.Context, r *http.Request, currentId int64) ([]apiSearch.Headline, interface{}, int, int, error) {
	// Get the details of the current user
	feeds, err := GetFeedsByResourceId(c, r, "ContactId", currentId)
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

func GetHeadlinesForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return GetHeadlinesForContactById(c, r, currentId)
}

func GetFeedForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ContactId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	feed, total, err := apiSearch.SearchFeedForContacts(c, r, []models.Contact{contact}, feeds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return feed, nil, len(feed), total, nil
}

func GetFeedsForContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	feeds, err := GetFeedsByResourceId(c, r, "ContactId", currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, 0, err
	}

	return feeds, nil, len(feeds), 0, nil
}

func FilterContacts(c context.Context, r *http.Request, queryType, query string) ([]models.Contact, error) {
	// User has to be logged in
	_, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		return []models.Contact{}, err
	}

	return filterContacts(c, r, queryType, query)
}

/*
* Create methods
 */

func CreateContact(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, int, error) {
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
			return []models.Contact{}, nil, 0, 0, err
		}

		newContacts := []models.Contact{}
		for i := 0; i < len(contacts); i++ {
			_, err = createContact(c, r, &contacts[i])
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Contact{}, nil, 0, 0, err
			}
			newContacts = append(newContacts, contacts[i])
		}

		includes := getIncludesForContacts(c, r, newContacts)
		mediaList, err := getMediaList(c, r, contacts[0].ListId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return newContacts, includes, len(newContacts), 0, nil
		}

		contactsWithDefaults, err := ContactsToDefaultFields(c, r, newContacts, mediaList)
		return contactsWithDefaults, includes, len(contactsWithDefaults), 0, nil
	}

	// Create contact
	_, err = createContact(c, r, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	contacts := []models.Contact{contact}
	includes := getIncludesForContacts(c, r, contacts)
	mediaList, err := getMediaList(c, r, contacts[0].ListId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return contacts, includes, len(contacts), 0, nil
	}

	contactsWithDefaults, err := ContactsToDefaultFields(c, r, contacts, mediaList)
	return contactsWithDefaults, includes, 0, 0, nil
}

// Does a ES sync in parse package & Twitter sync here
func BatchCreateContactsForDuplicateList(c context.Context, r *http.Request, contacts []models.Contact, mediaListId int64) ([]int64, error) {
	var previousKeys []int64
	var keys []*datastore.Key
	var contactIds []int64

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, err
	}

	for i := 0; i < len(contacts); i++ {
		// Previous keys of contacts so we can get lists later
		previousKeys = append(previousKeys, contacts[i].Id)

		// Remove list specific features for a contact
		contacts[i].Id = 0
		contacts[i].CreatedBy = currentUser.Id
		contacts[i].Created = time.Now()
		contacts[i].Updated = time.Now()
		contacts[i].ListId = mediaListId
		contacts[i].Normalize()
		keys = append(keys, contacts[i].Key(c))
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
		// Duplicate Feed
		feeds, err := GetFeedsByResourceId(c, r, "ContactId", previousKeys[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return []int64{}, err
		}

		for i := 0; i < len(feeds); i++ {
			feeds[i].Id = 0
			feeds[i].ListId = mediaListId
			feeds[i].ContactId = ks[i].IntID()
			feeds[i].Create(c, r, currentUser)
		}

		contactIds = append(contactIds, ks[i].IntID())
	}

	return contactIds, nil
}

// Does a ES sync in parse package & Twitter sync here
func BatchCreateContactsForExcelUpload(c context.Context, r *http.Request, contacts []models.Contact, mediaListId int64) ([]int64, []int64, error) {
	var keys []*datastore.Key
	var contactIds []int64
	var publicationIds []int64
	var selectedContacts []models.Contact

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, []int64{}, err
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].CreatedBy = currentUser.Id
		contacts[i].Created = time.Now()
		contacts[i].Updated = time.Now()
		contacts[i].ListId = mediaListId
		contacts[i].Normalize()
		contacts[i].FormatName()
		keys = append(keys, contacts[i].Key(c))

		for x := 0; x < len(contacts[i].Employers); x++ {
			publicationIds = append(publicationIds, contacts[i].Employers[x])
		}

		for x := 0; x < len(contacts[i].PastEmployers); x++ {
			publicationIds = append(publicationIds, contacts[i].PastEmployers[x])
		}

		selectedContacts = append(selectedContacts, contacts[i])
	}

	ks := []*datastore.Key{}

	err = nds.RunInTransaction(c, func(ctx context.Context) error {
		contextWithTimeout, _ := context.WithTimeout(c, time.Second*1000)
		ks, err = nds.PutMulti(contextWithTimeout, keys, selectedContacts)
		if err != nil {
			log.Errorf(c, "%v", err)
			return err
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf(c, "%v", err)
		return []int64{}, []int64{}, err
	}

	for i := 0; i < len(ks); i++ {
		contactIds = append(contactIds, ks[i].IntID())
	}

	return contactIds, publicationIds, nil
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func Save(c context.Context, r *http.Request, ct *models.Contact) (*models.Contact, error) {
	ct.Normalize()

	if ct.Email != "" && len(ct.Employers) == 0 {
		contactURLArray := strings.Split(ct.Email, "@")
		companyData, err := apiSearch.SearchCompanyDatabase(c, r, contactURLArray[1])
		if err == nil {
			isEmailProvider := false

			for i := 0; i < len(companyData.Data.Category); i++ {
				if companyData.Data.Category[i].Code == "EMAIL_PROVIDER" {
					isEmailProvider = true
				}
			}

			if !isEmailProvider {
				trimPublicationName := strings.Trim(companyData.Data.Organization.Name, " ")
				if trimPublicationName != "" {
					publication, err := UploadFindOrCreatePublication(c, r, companyData.Data.Organization.Name, companyData.Data.Website)
					if err == nil {
						ct.Employers = append(ct.Employers, publication.Id)
					} else {
						log.Errorf(c, "%v", err)
						log.Infof(c, "%v", companyData)
					}
				}
			}
		} else {
			log.Errorf(c, "%v", err)
			log.Infof(c, "%v", contactURLArray)
		}
	}

	ct.Normalize()
	ct.Save(c, r)
	sync.ResourceSync(r, ct.Id, "Contact", "create")
	return ct, nil
}

func EnrichContact(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get the details of the current contact
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, errors.New("Could not get user")
	}

	mediaList, err := getMediaList(c, r, contact.ListId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	if mediaList.TeamId != user.TeamId && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
		return models.Contact{}, nil, errors.New("You don't have permissions to edit these objects")
	}

	_, err = enrichContact(c, r, &contact)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	_, err = Save(c, r, &contact)

	// Sync with ES
	sync.ResourceSync(r, contact.Id, "Contact", "create")

	// If user is just created
	if contact.Twitter != "" {
		sync.TwitterSync(r, contact.Twitter)
	}
	if contact.Instagram != "" {
		sync.InstagramSync(r, contact.Instagram, user.InstagramAuthKey)
	}

	return contact, nil, nil
}

func UpdateSingleContact(c context.Context, r *http.Request, id string) (models.Contact, interface{}, error) {
	// Get the details of the current contact
	contact, _, err := GetContact(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, errors.New("Could not get user")
	}

	mediaList, err := getMediaList(c, r, contact.ListId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Contact{}, nil, err
	}

	if mediaList.TeamId != user.TeamId && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
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

func UpdateBatchContact(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedContacts []models.Contact
	err := decoder.Decode(buf, &updatedContacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	// Get logged in user
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, errors.New("Could not get user")
	}

	// Check if each of the contacts have permissions before updating anything
	currentContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		contact, err := getContact(c, r, updatedContacts[i].Id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, 0, err
		}

		mediaList, err := getMediaList(c, r, contact.ListId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, 0, err
		}

		if mediaList.TeamId != user.TeamId && !permissions.AccessToObject(contact.CreatedBy, user.Id) && !user.IsAdmin {
			return []models.Contact{}, nil, 0, 0, errors.New("Forbidden")
		}

		currentContacts = append(currentContacts, contact)
	}

	// Update each of the contacts
	newContacts := []models.Contact{}
	for i := 0; i < len(updatedContacts); i++ {
		updatedContact, _, err := updateContact(c, r, &currentContacts[i], updatedContacts[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Contact{}, nil, 0, 0, err
		}

		newContacts = append(newContacts, updatedContact)
	}

	return newContacts, nil, len(newContacts), 0, nil
}

func CopyContacts(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, int, error) {
	// Get logged in user
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, errors.New("Could not get user")
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var copyContacts copyContactsDetails
	err = decoder.Decode(buf, &copyContacts)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, err
	}

	newContacts := []models.Contact{}
	newContactIds := []int64{}

	// Add contact to the other media list
	mediaList, err := getMediaListBasic(c, r, copyContacts.ListId)
	if err != nil {
		return []models.Contact{}, nil, 0, 0, err
	}

	mediaListFields := map[string]bool{}
	for i := 0; i < len(mediaList.FieldsMap); i++ {
		if mediaList.FieldsMap[i].CustomField && !mediaList.FieldsMap[i].ReadOnly {
			if _, ok := mediaListFields[mediaList.FieldsMap[i].Value]; !ok {
				mediaListFields[mediaList.FieldsMap[i].Value] = true
			}
		}
	}

	for i := 0; i < len(copyContacts.Contacts); i++ {
		contact, err := getContact(c, r, copyContacts.Contacts[i])
		if err == nil {
			previousContactId := contact.Id
			contact.Id = 0

			contact.CreatedBy = user.Id
			contact.Created = time.Now()
			contact.Updated = time.Now()

			previousCustomFields := contact.CustomFields
			contact.CustomFields = []models.CustomContactField{}

			for x := 0; x < len(previousCustomFields); x++ {
				customFieldName := previousCustomFields[x].Name
				if _, ok := mediaListFields[customFieldName]; ok {
					contact.CustomFields = append(contact.CustomFields, previousCustomFields[x])
				}
			}

			log.Infof(c, "%v", contact.CustomFields)

			contact.ListId = copyContacts.ListId
			contact.Normalize()
			contact.Create(c, r, user)

			newContactIds = append(newContactIds, contact.Id)
			newContacts = append(newContacts, contact)

			// Copy all of their feeds
			feeds, err := GetFeedsByResourceId(c, r, "ContactId", previousContactId)
			if err != nil {
				log.Errorf(c, "%v", err)
				return nil, nil, 0, 0, err
			}

			for x := 0; x < len(feeds); x++ {
				feeds[x].Id = 0
				feeds[x].CreatedBy = user.Id
				feeds[x].Created = time.Now()
				feeds[x].Updated = time.Now()

				feeds[x].ContactId = contact.Id
				feeds[x].ListId = copyContacts.ListId

				feeds[x].Create(c, r, user)
			}
		}
	}

	// Append media list
	mediaList.Contacts = append(mediaList.Contacts, newContactIds...)
	mediaList.Save(c)

	// Sync all the contacts in bulk here
	sync.ListUploadResourceBulkSync(r, mediaList.Id, mediaList.Contacts, []int64{})
	return newContacts, nil, 0, 0, nil
}

/*
* Delete methods
 */

func BulkDeleteContacts(c context.Context, r *http.Request) ([]models.Contact, interface{}, int, int, error) {
	// Get logged in user
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, nil, 0, 0, errors.New("Could not get user")
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var deleteContacts deleteContactsDetails
	err = decoder.Decode(buf, &deleteContacts)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	if len(deleteContacts.Contacts) > 0 {
		var listId int64

		contacts := []models.Contact{}
		contactIds := []int64{}
		for i := 0; i < len(deleteContacts.Contacts); i++ {
			contact, err := getContact(c, r, deleteContacts.Contacts[i])
			if err == nil {
				if contact.CreatedBy == user.Id {
					if contact.ListId != 0 {
						listId = contact.ListId
					}

					contact.IsDeleted = true
					contact.Save(c, r)

					contactIds = append(contactIds, contact.Id)
					contacts = append(contacts, contact)
				}
			}
		}

		sync.ListUploadResourceBulkSync(r, listId, contactIds, []int64{})
		return contacts, nil, len(contacts), 0, nil
	}

	return []models.Contact{}, nil, 0, 0, nil
}

func DeleteContact(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	// Update contact
	contact, err := getContact(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	mediaList, err := getMediaList(c, r, contact.ListId)
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
	if mediaList.TeamId != user.TeamId && !permissions.AccessToObject(contact.CreatedBy, user.Id) {
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
