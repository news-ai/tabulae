package incoming

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/julienschmidt/httprouter"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/notifications"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/errors"
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

func InternalTrackerHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	for i := 0; i < len(allEvents); i++ {
		singleEvent := allEvents[i]
		if singleEvent.SgMessageID == "" {
			email, _, err := controllers.GetEmail(c, r, singleEvent.ID)
			emailIds = append(emailIds, email.Id)
			notification := models.NotificationChange{}

			// If there is an error
			if err != nil {
				log.Errorf(c, "%v", err)
				errors.ReturnError(w, http.StatusInternalServerError, "Internal Tracker issue", err.Error())
				return
			}

			// Add to appropriate Email model
			switch singleEvent.Event {
			case "open":
				for x := 0; x < singleEvent.Count; x++ {
					_, notification, err = controllers.MarkOpened(c, r, &email)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", singleEvent)
						log.Errorf(c, "%v", err)
					}
				}
			case "click":
				for x := 0; x < singleEvent.Count; x++ {
					_, notification, err = controllers.MarkClicked(c, r, &email)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", singleEvent)
						log.Errorf(c, "%v", err)
					}
				}
			default:
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
			}

			// Send user notification
			if notification.Verb != "" {
				// Send the notification to the user if they have a socket open
				currentUser, err := controllers.GetCurrentUser(c, r)
				if err != nil {
					log.Errorf(c, "%v", err)
					w.WriteHeader(500)
					return
				}

				notificationChanges := []models.NotificationChange{}
				notificationChanges = append(notificationChanges, notification)
				notifications.SendNotification(r, notificationChanges, currentUser.Id)
			}
		} else {
			// Validate email exists with particular SendGridId
			sendGridId := strings.Split(singleEvent.SgMessageID, ".")[0]
			email, err := controllers.FilterEmailBySendGridID(c, sendGridId)
			emailIds = append(emailIds, email.Id)
			notification := models.NotificationChange{}
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
				log.Errorf(c, "%v with value %v", err, sendGridId)
			}

			// Add to appropriate Email model
			switch singleEvent.Event {
			case "bounce", "dropped":
				_, notification, err = controllers.MarkBounced(c, r, &email, singleEvent.Reason)
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
				_, notification, err = controllers.MarkSpam(c, r, &email)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			default:
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
			}

			// Send user notification
			if notification.Verb != "" {
				// Send the notification to the user if they have a socket open
				currentUser, err := controllers.GetCurrentUser(c, r)
				if err != nil {
					log.Errorf(c, "%v", err)
					w.WriteHeader(500)
					return
				}

				notificationChanges := []models.NotificationChange{}
				notificationChanges = append(notificationChanges, notification)
				notifications.SendNotification(r, notificationChanges, currentUser.Id)
			}
		}
	}

	if hasErrors {
		errors.ReturnError(w, http.StatusInternalServerError, "Internal Tracker handling error", "Problem parsing data")
		return
	}

	sync.EmailResourceBulkSync(r, emailIds)
	w.WriteHeader(200)
	return
}
