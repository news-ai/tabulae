package incoming

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
)

type SendGridEvent struct {
	SgMessageID string `json:"sg_message_id"`
	Email       string `json:"email"`
	Timestamp   int    `json:"timestamp"`
	Event       string `json:"event"`
}

type BounceEvent struct {
	Status       string   `json:"status"`
	SgEventID    string   `json:"sg_event_id"`
	SgMessageID  string   `json:"sg_message_id"`
	Event        string   `json:"event"`
	Email        string   `json:"email"`
	Timestamp    int      `json:"timestamp"`
	SMTPID       string   `json:"smtp-id"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	AsmGroupID int    `json:"asm_group_id"`
	Reason     string `json:"reason"`
	Type       string `json:"type"`
	IP         string `json:"ip"`
	TLS        string `json:"tls"`
	CertErr    string `json:"cert_err"`
}

type ClickEvent struct {
	SgEventID   string `json:"sg_event_id"`
	SgMessageID string `json:"sg_message_id"`
	IP          string `json:"ip"`
	Useragent   string `json:"useragent"`
	Event       string `json:"event"`
	Email       string `json:"email"`
	Timestamp   int    `json:"timestamp"`
	URL         string `json:"url"`
	URLOffset   struct {
		Index int    `json:"index"`
		Type  string `json:"type"`
	} `json:"url_offset"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	AsmGroupID int `json:"asm_group_id"`
}

type DeliveredEvent struct {
	Response     string   `json:"response"`
	SgEventID    string   `json:"sg_event_id"`
	SgMessageID  string   `json:"sg_message_id"`
	Event        string   `json:"event"`
	Email        string   `json:"email"`
	Timestamp    int      `json:"timestamp"`
	SMTPID       string   `json:"smtp-id"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	AsmGroupID int    `json:"asm_group_id"`
	IP         string `json:"ip"`
	TLS        string `json:"tls"`
	CertErr    string `json:"cert_err"`
}

type OpenEvent struct {
	Email        string   `json:"email"`
	Timestamp    int      `json:"timestamp"`
	IP           string   `json:"ip"`
	SgEventID    string   `json:"sg_event_id"`
	SgMessageID  string   `json:"sg_message_id"`
	Useragent    string   `json:"useragent"`
	Event        string   `json:"event"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	AsmGroupID int `json:"asm_group_id"`
}

func SendGridHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		c := appengine.NewContext(r)

		buf, _ := ioutil.ReadAll(r.Body)

		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
		decoder := json.NewDecoder(rdr1)
		var allEvents []SendGridEvent
		err := decoder.Decode(&allEvents)

		// If there is an error
		if err != nil {
			log.Errorf(c, "%v", err)
			permissions.ReturnError(w, http.StatusInternalServerError, "SendGrid issue", err.Error())
			return
		}

		for i := 0; i < len(allEvents); i++ {
			singleEvent := allEvents[i]

			// Validate email exists with particular SendGridId
			email, err := controllers.FilterEmailBySendGridID(c, singleEvent.SgMessageID)
			if err != nil {
				log.Errorf(c, "%v", err)
			}

			// Another decoder variable
			rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
			eventDecoder := json.NewDecoder(rdr2)

			// Add to appropriate Email model
			switch singleEvent.Event {
			case "bounce":
				var bounceEvent BounceEvent
				err = eventDecoder.Decode(&bounceEvent)
				if err != nil {
					log.Errorf(c, "%v", err)
				}

				_, err = email.MarkBounced(c, bounceEvent.Reason)
				if err != nil {
					log.Errorf(c, "%v", err)
				}
			case "click":
				var clickEvent ClickEvent
				err := eventDecoder.Decode(&clickEvent)
				if err != nil {
					log.Errorf(c, "%v", err)
				}

				_, err = email.MarkClicked(c)
				if err != nil {
					log.Errorf(c, "%v", err)
				}
			case "delivered":
				var delieveredEvent DeliveredEvent
				err := eventDecoder.Decode(&delieveredEvent)
				if err != nil {
					log.Errorf(c, "%v", err)
				}

				_, err = email.MarkDelivered(c)
				if err != nil {
					log.Errorf(c, "%v", err)
				}
			case "open":
				var openEvent OpenEvent
				err := eventDecoder.Decode(&openEvent)
				if err != nil {
					log.Errorf(c, "%v", err)
				}

				_, err = email.MarkOpened(c)
				if err != nil {
					log.Errorf(c, "%v", err)
				}
			}
		}
	}

	permissions.ReturnError(w, http.StatusInternalServerError, "SendGrid handling error", "method not implemented")
	return
}
