package updates

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/errors"
	"github.com/news-ai/web/utilities"
)

type InternalTrackerEvent struct {
	// Similar between both
	Event string `json:"event"`

	// Internal tracker
	ID    string `json:"id"`
	Count int    `json:"count"`

	// Sendgrid data
	SgMessageID string `json:"sg_message_id"`
	Email       string `json:"email"`
	Timestamp   int    `json:"timestamp"`
	Reason      string `json:"reason"`

	// Sendgrid<->Tabulae data
	EmailId   string `json:"emailId"`
	CreatedBy string `json:"createdBy"`
}

func internalTrackerHandler(w http.ResponseWriter, r *http.Request) {
	hasErrors := false
	c := appengine.NewContext(r)

	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	decoder := json.NewDecoder(rdr1)
	var allEvents []InternalTrackerEvent
	err := decoder.Decode(&allEvents)

	// If there is an error
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Internal Tracker issue", err.Error())
		return
	}

	emailIdsDatastoreMap := map[int64]bool{}
	for i := 0; i < len(allEvents); i++ {
		// Find which emailId to convert
		idToConvert := ""
		if allEvents[i].SgMessageID == "" {
			idToConvert = allEvents[i].ID
		} else {
			if allEvents[i].EmailId != "" {
				idToConvert = allEvents[i].EmailId
			}
		}

		// Convert id to int64
		if idToConvert != "" {
			emailId, err := utilities.StringIdToInt(idToConvert)
			if err != nil {
				log.Errorf(c, "%v", err)
				continue
			}
			emailIdsDatastoreMap[emailId] = true
		}
	}

	// Convert keys in map to int64 ids array
	emailIdsDatastore := []int64{}
	for k := range emailIdsDatastoreMap {
		emailIdsDatastore = append(emailIdsDatastore, k)
	}

	datastoreEmails, _, err := controllers.GetEmailUnauthorizedBulk(c, r, emailIdsDatastore)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Updates handing error", err.Error())
		return
	}

	// Convert datastore emails to a dictionary for easy lookup
	emailIdToEmail := map[int64]models.Email{}
	for i := 0; i < len(datastoreEmails); i++ {
		emailIdToEmail[datastoreEmails[i].Id] = datastoreEmails[i]
	}

	for i := 0; i < len(allEvents); i++ {
		singleEvent := allEvents[i]
		var emailId int64
		var err error
		var email models.Email

		if singleEvent.SgMessageID == "" {
			emailId, err = utilities.StringIdToInt(singleEvent.ID)
			if err != nil {
				hasErrors = true
				log.Debugf(c, "%v", singleEvent)
				log.Errorf(c, "%v", err)
				continue
			}
			email = emailIdToEmail[emailId]
		} else {
			if singleEvent.EmailId != "" {
				log.Infof(c, "%v", singleEvent.EmailId)
				emailId, err := utilities.StringIdToInt(singleEvent.EmailId)
				if err != nil {
					hasErrors = true
					log.Debugf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
					continue
				}
				email = emailIdToEmail[emailId]
			} else {
				// Validate email exists with particular SendGridId
				sendGridId := strings.Split(singleEvent.SgMessageID, ".")[0]
				email, err := controllers.FilterEmailBySendGridID(c, sendGridId)

				// Check if there's any errors
				if err != nil {
					hasErrors = true
					log.Debugf(c, "%v", singleEvent)
					log.Errorf(c, "%v with value %v", err, sendGridId)
					continue
				}

				// Set the email sendgrid id
				emailId = email.Id
				email = emailIdToEmail[emailId]
				email.SendGridId = sendGridId
			}
		}

		if singleEvent.SgMessageID != "" {
			// Sendgrid event
			switch singleEvent.Event {
			case "bounce":
				email.BouncedReason = singleEvent.Reason
				email.Delievered = true
				email.Bounced = true
			case "delivered":
				email.Delievered = true
			case "spamreport":
				email.Delievered = true
				email.Spam = true
			case "open":
				email.Delievered = true
				email.SendGridOpened += 1
			case "dropped":
				email.Delievered = true
				email.Dropped = true
			default:
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
			}
		} else {
			// Track event
			switch singleEvent.Event {
			case "open":
				for x := 0; x < singleEvent.Count; x++ {
					email.Opened += 1
				}
			case "click":
				for x := 0; x < singleEvent.Count; x++ {
					email.Clicked += 1
				}
			case "unsubscribe":
				if email.To != "" {
					unsubscribe := models.ContactUnsubscribe{}
					unsubscribe.CreatedBy = email.CreatedBy
					unsubscribe.ListId = email.ListId
					unsubscribe.ContactId = email.ContactId

					unsubscribe.Email = email.To
					unsubscribe.Unsubscribed = true
					_, err = unsubscribe.Create(c, r)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", err)
					}
				}
			default:
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
			}
		}

		emailIdToEmail[emailId] = email
	}

	var keys []*datastore.Key
	updatedEmails := []models.Email{}
	emailIds := []int64{}
	memcacheKeys := []string{}
	for _, email := range emailIdToEmail {
		// Invalidate memcache for this particular campaign
		memcacheKey := controllers.GetEmailCampaignKey(email)
		memcacheKeys = append(memcacheKeys, memcacheKey)

		emailIds = append(emailIds, email.Id)

		keys = append(keys, email.Key(c))
		updatedEmails = append(updatedEmails, email)
	}

	var ks []*datastore.Key
	err = nds.RunInTransaction(c, func(ctx context.Context) error {
		contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
		ks, err = nds.PutMulti(contextWithTimeout, keys, updatedEmails)
		if err != nil {
			log.Errorf(c, "%v", err)
			return err
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf(c, "%v", err)
	}

	// Even if error let's sync the data correctly first
	if len(memcacheKeys) > 0 {
		noDuplicatesMemcache := utilities.RemoveDuplicatesUnordered(memcacheKeys)
		log.Infof(c, "%v", noDuplicatesMemcache)
		err = memcache.DeleteMulti(c, noDuplicatesMemcache)
		if err != nil {
			log.Warningf(c, "%v", err)
		}
	}

	if len(emailIds) > 0 {
		sync.EmailResourceBulkSync(r, emailIds)
	}

	if hasErrors {
		errors.ReturnError(w, http.StatusInternalServerError, "Internal Tracker handling error", "Problem parsing data")
		return
	}

	w.WriteHeader(200)
	return
}
