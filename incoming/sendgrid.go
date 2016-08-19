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
	"github.com/news-ai/tabulae/permissions"
)

type SendGridEvent struct {
	SgMessageID string `json:"sg_message_id"`
	Email       string `json:"email"`
	Timestamp   int    `json:"timestamp"`
	Event       string `json:"event"`
	Reason      string `json:"reason"`
}

func SendGridHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	switch r.Method {
	case "POST":
		hasErrors := false
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
			sendGridId := strings.Split(singleEvent.SgMessageID, ".")[0]
			email, err := controllers.FilterEmailBySendGridID(c, sendGridId)
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", singleEvent)
				log.Errorf(c, "%v with value %v", err, sendGridId)
			}

			// Add to appropriate Email model
			switch singleEvent.Event {
			case "bounce":
				_, err = email.MarkBounced(c, singleEvent.Reason)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "click":
				_, err = email.MarkClicked(c)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "delivered":
				_, err = email.MarkDelivered(c)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", singleEvent)
					log.Errorf(c, "%v", err)
				}
			case "open":
				_, err = email.MarkOpened(c)
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

		if hasErrors {
			permissions.ReturnError(w, http.StatusInternalServerError, "SendGrid handling error", "Problem parsing data")
			return
		}
		w.WriteHeader(200)
		return
	}

	permissions.ReturnError(w, http.StatusInternalServerError, "SendGrid handling error", "method not implemented")
	return
}
