package updates

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"

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
}

func formatCampaignName(campaignName string) string {
	campaignName = utilities.RemoveSpecialCharacters(campaignName)
	campaignName = strings.ToLower(campaignName)
	campaignName = strings.Trim(campaignName, " ")
	campaignName = strings.Replace(campaignName, " ", "-", -1)
	return campaignName
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

	emailIds := []int64{}
	memcacheKeys := []string{}

	for i := 0; i < len(allEvents); i++ {
		singleEvent := allEvents[i]
		if singleEvent.SgMessageID == "" {
			email, _, err := controllers.GetEmailUnauthorized(c, r, singleEvent.ID)
			emailIds = append(emailIds, email.Id)

			// If there is an error
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
				log.Errorf(c, "%v", err)
				errors.ReturnError(w, http.StatusInternalServerError, "Internal Tracker issue", err.Error())
				continue
			}

			// Add to appropriate Email model
			switch singleEvent.Event {
			case "open":
				for x := 0; x < singleEvent.Count; x++ {
					_, err = controllers.MarkOpened(c, r, &email)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", singleEvent)
						log.Errorf(c, "%v", err)
					}
				}
			case "click":
				for x := 0; x < singleEvent.Count; x++ {
					_, err = controllers.MarkClicked(c, r, &email)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", singleEvent)
						log.Errorf(c, "%v", err)
					}
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

			emailSubject := email.Subject
			if email.BaseSubject != "" {
				emailSubject = email.BaseSubject
			}

			// Invalidate memcache for this particular campaign
			userIdString := strconv.FormatInt(email.CreatedBy, 10)
			dayFormat := email.Created.Format("2006-01-02")
			campaignName := formatCampaignName(emailSubject)
			memcacheKey := userIdString + "-" + dayFormat + "-" + campaignName
			memcacheKeys = append(memcacheKeys, memcacheKey)
		} else {
			// Validate email exists with particular SendGridId
			sendGridId := strings.Split(singleEvent.SgMessageID, ".")[0]
			email, err := controllers.FilterEmailBySendGridID(c, sendGridId)
			emailIds = append(emailIds, email.Id)
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
				log.Errorf(c, "%v with value %v", err, sendGridId)
				continue
			}

			// Add to appropriate Email model
			// https://sendgrid.com/docs/API_Reference/Webhooks/event.html
			switch singleEvent.Event {
			case "bounce":
				_, err = controllers.MarkBounced(c, r, &email, singleEvent.Reason)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "delivered":
				_, err = controllers.MarkDelivered(c, r, &email)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "spamreport":
				_, err = controllers.MarkSpam(c, r, &email)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "open":
				_, err = controllers.MarkSendgridOpen(c, r, &email)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "dropped":
				_, err = controllers.MarkSendgridDrop(c, r, &email)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			default:
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
			}
		}
	}

	if hasErrors {
		errors.ReturnError(w, http.StatusInternalServerError, "Internal Tracker handling error", "Problem parsing data")
		return
	}

	if len(memcacheKeys) > 0 {
		noDuplicatesMemcache := utilities.RemoveDuplicatesUnordered(memcacheKeys)
		log.Infof(c, "%v", noDuplicatesMemcache)
		err = memcache.DeleteMulti(c, noDuplicatesMemcache)
		if err != nil {
			log.Errorf(c, "%v", err)
		}
	}

	if len(emailIds) > 0 {
		sync.EmailResourceBulkSync(r, emailIds)
	}

	w.WriteHeader(200)
	return
}
