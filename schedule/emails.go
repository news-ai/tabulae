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

		files := []models.File{}
		if len(schedueled[i].Attachments) > 0 {
			for x := 0; x < len(schedueled[i].Attachments); x++ {
				file, _, err := controllers.GetFileByIdUnauthorized(c, r, schedueled[i].Attachments[x])
				if err == nil {
					files = append(files, file)
				} else {
					hasErrors = true
					log.Errorf(c, "%v", err)
				}
			}
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

				log.Infof(c, "%v", files)
				gmailId, gmailThreadId, err := emails.SendGmailEmail(r, user, schedueled[i], files)
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

				// if files are present
				for i := 0; i < len(files); i++ {
					files[i].Imported = true
					files[i].Save(c)
				}
			}
		} else {
			emailSent, emailId, err := emails.SendEmail(r, schedueled[i], user, files)
			if err != nil {
				hasErrors = true
				log.Errorf(c, "%v", err)
				continue
			}

			schedueled[i].Save(c)

			if emailSent {
				// Set attachments for deletion
				for i := 0; i < len(files); i++ {
					files[i].Imported = true
					files[i].Save(c)
				}

				_, err := schedueled[i].MarkSent(c, emailId)
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
