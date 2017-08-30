package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	apiControllers "github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/controllers"
	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/emails"
	"github.com/news-ai/web/google"
	"github.com/news-ai/web/outlook"
)

func getUser(c context.Context, id int64) (apiModels.User, error) {
	// Get the current signed in user details by Id
	var user apiModels.User
	userId := datastore.NewKey(c, "User", "", id, nil)
	err := datastoreClient.Get(c, userId, &user)

	if err != nil {
		log.Errorf(c, "%v", err)
		return apiModels.User{}, err
	}

	if user.Email != "" {
		user.Format(userId, "users")
		return user, nil
	}
	return apiModels.User{}, errors.New("No user by this id")
}

func getCurrentSchedueledEmails(c context.Context) ([]models.Email, error) {
	emails := []models.Email{}

	timeNow := time.Now()
	currentTime := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), timeNow.Hour(), timeNow.Minute(), 0, 0, time.FixedZone("GMT", 0))

	// When the email is "Delievered == false" and has a "SendAt" date
	// And "Cancel == false". Also a filer if the user has sent it already or not
	query := datastore.NewQuery("Email").Filter("SendAt <=", currentTime).Filter("IsSent =", true).Filter("Delievered =", false).Filter("Cancel =", false)

	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	emails = make([]models.Email, len(ks))
	if err := datastoreClient.GetMulti(c, keys, emails); err != nil {
		log.Errorf(c, "%v", err)
		return []models.Email{}, err
	}

	emailsToSend := []models.Email{}
	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")

		if !emails[i].SendAt.IsZero() {
			emailsToSend = append(emailsToSend, emails[i])
		}
	}

	return emailsToSend, nil
}

func SchedueleEmailTask(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	hasErrors := false
	schedueled, err := getCurrentSchedueledEmails(r)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	log.Infof(c, "%v", len(schedueled))

	// Loop through the emails and send them
	emailIds := []int64{}
	for i := 0; i < len(schedueled); i++ {
		user, err := getUser(c, schedueled[i].CreatedBy)
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
				log.Infof(c, "%v", schedueled[i])
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

				// Add to emailids array
				emailIds = append(emailIds, schedueled[i].Id)
			}
		} else if schedueled[i].Method == "outlook" {
			if user.OutlookAccessToken != "" && user.Outlook {
				err = outlook.ValidateAccessToken(r, user)
				// Refresh access token if err is nil
				if err != nil {
					log.Errorf(c, "%v", err)
					user, err = outlook.RefreshAccessToken(r, user)
					if err != nil {
						hasErrors = true
						log.Errorf(c, "%v", err)
						continue
					}
				}

				log.Infof(c, "%v", files)
				log.Infof(c, "%v", schedueled[i])
				err := emails.SendOutlookEmail(r, user, schedueled[i], files)
				if err != nil {
					hasErrors = true
					log.Errorf(c, "%v", err)
					continue
				}

				_, err = schedueled[i].MarkDelivered(c)
				// sync.ResourceSync(r, schedueled[i].Id, "Email", "create")
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

				// Add to emailids array
				emailIds = append(emailIds, schedueled[i].Id)
			}
		} else if schedueled[i].Method == "smtp" {
			// If scheduled from SMTP.
			// If their SMTP is valid & they are using an external email.
			// We assume all this information is correct & they can send
			// emails through SMTP.
			if user.SMTPValid && user.ExternalEmail && user.EmailSetting != 0 {
				emailBody, err := emails.GenerateEmail(r, user, schedueled[i], files)
				if err != nil {
					log.Errorf(c, "%v", err)
					hasErrors = true
				}

				emailSetting, err := controllers.GetEmailSettingById(c, r, user.EmailSetting)
				if err != nil {
					log.Errorf(c, "%v", err)
					hasErrors = true
				}

				SMTPPassword := string(user.SMTPPassword[:])

				contextWithTimeout, _ := context.WithTimeout(c, time.Second*30)
				client := urlfetch.Client(contextWithTimeout)
				getUrl := "https://tabulae-smtp.newsai.org/send"

				sendEmailRequest := models.SMTPEmailSettings{}
				sendEmailRequest.Servername = emailSetting.SMTPServer + ":" + strconv.Itoa(emailSetting.SMTPPortSSL)
				sendEmailRequest.EmailUser = user.SMTPUsername
				sendEmailRequest.EmailPassword = SMTPPassword
				sendEmailRequest.To = schedueled[i].To
				sendEmailRequest.Subject = schedueled[i].Subject
				sendEmailRequest.Body = emailBody

				SendEmailRequest, err := json.Marshal(sendEmailRequest)
				if err != nil {
					log.Errorf(c, "%v", err)
					hasErrors = true
				}

				log.Infof(c, "%v", string(SendEmailRequest))
				sendEmailQuery := bytes.NewReader(SendEmailRequest)

				req, _ := http.NewRequest("POST", getUrl, sendEmailQuery)

				resp, err := client.Do(req)
				if err != nil {
					log.Errorf(c, "%v", err)
					hasErrors = true
				}
				defer resp.Body.Close()

				decoder := json.NewDecoder(resp.Body)
				var verifyResponse controllers.SMTPEmailResponse
				err = decoder.Decode(&verifyResponse)
				if err != nil {
					log.Errorf(c, "%v", err)
					hasErrors = true
				}

				log.Infof(c, "%v", verifyResponse)

				if verifyResponse.Status {
					_, err := schedueled[i].MarkDelivered(c)
					if err != nil {
						log.Errorf(c, "%v", err)
						hasErrors = true
					}
				} else {
					log.Errorf(c, "%v", verifyResponse)
					hasErrors = true
				}

				// Add to emailids array
				emailIds = append(emailIds, schedueled[i].Id)
			}
		} else {
			// If email is sent through SendGrid
			// and if it doesn't have a sendGridId
			// means that it hasn't been sent yet (pretty sure)
			// unless there was a freak event.
			if !schedueled[i].SendAt.IsZero() && schedueled[i].SendGridId == "" {
				log.Infof(c, "%v", schedueled[i])
				userBilling, _ := apiControllers.GetUserBilling(c, r, user)
				sendGridKey := emails.GetSendGridKeyForUser(userBilling)
				emailSent, emailId, err := emails.SendEmail(r, schedueled[i], user, files, sendGridKey)
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

					_, err = schedueled[i].MarkDelivered(c)
					if err != nil {
						log.Errorf(c, "%v", err)
						hasErrors = true
					}
				}

				// Add to emailids array
				emailIds = append(emailIds, schedueled[i].Id)
			}
		}
	}

	// Sync the emails we have delivered
	sync.EmailResourceBulkSync(r, emailIds)

	if hasErrors {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	return
}
