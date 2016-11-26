package schedule

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/controllers"
	"google.golang.org/appengine/log"

	"github.com/news-ai/web/emails"
	"github.com/news-ai/web/google"
)

// When the email is "Delievered == false" and has a "SendAt" date
// And "Cancel == false"

func SchedueleEmailTask(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	hasErrors := false

	schedueled, err := controllers.GetCurrentSchedueledEmails(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	log.Infof(c, "%v", len(schedueled))

	// Loop through the emails and send them
	for i := 0; i < len(schedueled); i++ {
		if schedueled[i].Method == "gmail" {
			user, err := controllers.GetUserByIdUnauthorized(c, r, schedueled[i].CreatedBy)
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", err)
				continue
			}

			if user.AccessToken != "" && user.Gmail {
				err = google.ValidateAccessToken(r, user)
				// Refresh access token if err is nil
				if err != nil {
					log.Errorf(c, "%v", err)
					user, err = google.RefreshAccessToken(r, user)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", err)
						continue
					}
				}

				gmailId, gmailThreadId, err := emails.SendGmailEmail(r, user, schedueled[i])
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", err)
					continue
				}

				schedueled[i].GmailId = gmailId
				schedueled[i].GmailThreadId = gmailThreadId

				_, err = schedueled[i].MarkDelivered(c)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", err)
					continue
				}
			}
		}
	}

	if hasErrors {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	return
}
