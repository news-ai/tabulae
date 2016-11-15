package tasks

import (
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/errors"
)

func MakeUsersInactive(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	users, err := controllers.GetUsersUnauthorized(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get users", err.Error())
		return
	}

	for i := 0; i < len(users); i++ {
		billing, err := controllers.GetUserBilling(c, r, users[i])
		if err != nil {
			log.Errorf(c, "%v", err)
			continue
		}

		// For now only consider when they are on trial
		if billing.IsOnTrial {
			if billing.Expires.Before(time.Now()) {
				users[i].IsActive = false
				users[i].Save(c)
				sync.ResourceSync(r, users[i].Id, "User", "create")

				billing.IsOnTrial = false
				billing.Save(c)
			} else {
				// If trial expiring email hasn't been sent
				if !billing.TrialEmailSent {
					// Check if it expires tomorrow
					tomorrow := time.Now().AddDate(0, 0, 1)
					expiresAt := billing.Expires
					difference := tomorrow.YearDay() - expiresAt.YearDay()

					if difference == 0 {
						// Send the email to the user alerting them that their subscription is going to expire
						emailTrialExpires, _ := controllers.CreateEmailInternal(r, users[i].Email, users[i].FirstName, users[i].LastName)
						emailSent, emailId, err := emails.SendTrialExpiresTomorrowEmail(r, emailTrialExpires)
						if !emailSent || err != nil {
							// Redirect user back to login page
							log.Errorf(c, "%v", "Trial expires email email was not sent for "+users[i].Email)
							log.Errorf(c, "%v", err)
						}
						emailTrialExpires.MarkSent(c, emailId)

						billing.TrialEmailSent = true
						billing.Save(c)
					}
				}
			}
		}
	}
}
