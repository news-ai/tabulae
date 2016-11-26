package schedule

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/controllers"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/models"

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
	log.Infof(c, "%v", schedueled)

	// Loop through the emails and send them
	for i := 0; i < len(schedueled); i++ {
		user, err := controllers.GetUserByIdUnauthorized(c, r, schedueled[i].CreatedBy)
		if err != nil {
			hasErrors = true
			log.Errorf(c, "%v", err)
			continue
		}

		if schedueled[i].Method == "gmail" {
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
		} else {
			files := []models.File{}
			if len(schedueled[i].Attachments) > 0 {
				for i := 0; i < len(schedueled[i].Attachments); i++ {
					file, err := controllers.GetFileById(c, r, schedueled[i].Attachments[i])
					if err == nil {
						files = append(files, file)
					} else {
						hasErrors = true
						log.Errorf(c, "%v", err)
					}
				}
			}

			emailSent, emailId, batchId, err := emails.SendEmail(r, schedueled[i], user, files)
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", err)
				continue
			}

			schedueled[i].BatchId = batchId
			schedueled[i].Save(c)

			if emailSent {
				// Set attachments for deletion
				for i := 0; i < len(files); i++ {
					files[i].Imported = true
					files[i].Save(c)
				}

				val, err := schedueled[i].MarkSent(c, emailId)
				if err != nil {
					log.Errorf(c, "%v", err)
					hasErrors = true
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
