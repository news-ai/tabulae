package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"

	apiModels "github.com/news-ai/api/models"
	tabulaeModels "github.com/news-ai/tabulae/models"
	updateService "github.com/news-ai/tabulae/updates-service"
)

var (
	pubsubClient    *pubsub.Client
	subscription    *pubsub.Subscription
	datastoreClient *datastore.Client
)

func sendChangesToUpdateService(c context.Context, updates []updateService.EmailSendUpdate) error {
	updatesRequest, err := json.Marshal(updates)
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	log.Printf("%v", string(updatesRequest))
	updatesQuery := bytes.NewReader(updatesRequest)

	req, _ := http.NewRequest("POST", "https://updates-dot-newsai-1166.appspot.com/updates", updatesQuery)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func getEmails(c context.Context, ids []int64) ([]tabulaeModels.Email, apiModels.User, apiModels.Billing, []tabulaeModels.File, error) {
	var keys []*datastore.Key
	for i := 0; i < len(ids); i++ {
		emailId := datastore.IDKey("Email", ids[i], nil)
		keys = append(keys, emailId)
	}

	emails := make([]tabulaeModels.Email, len(keys))
	if err := datastoreClient.GetMulti(c, keys, emails); err != nil {
		log.Printf("%v", err)
		return []tabulaeModels.Email{}, apiModels.User{}, apiModels.Billing{}, []tabulaeModels.File{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Id = keys[i].ID
	}

	// Remove emails that have already been delivered
	// Downward loop fixes problems that you have when deleting
	// elements in an array while looping through them.
	for i := len(emails) - 1; i >= 0; i-- {
		if emails[i].Delievered {
			emails = append(emails[:i], emails[i+1:]...)
		}
	}

	if len(emails) == 0 {
		err := errors.New("Missing emails")
		return []tabulaeModels.Email{}, apiModels.User{}, apiModels.Billing{}, []tabulaeModels.File{}, err
	}

	// Get files if there are attachments
	var files []tabulaeModels.File
	if len(emails[0].Attachments) > 0 {
		var fileKeys []*datastore.Key
		for i := 0; i < len(emails[0].Attachments); i++ {
			fileId := datastore.IDKey("File", emails[0].Attachments[i], nil)
			fileKeys = append(fileKeys, fileId)
		}

		files = make([]tabulaeModels.File, len(fileKeys))
		if err := datastoreClient.GetMulti(c, fileKeys, files); err != nil {
			log.Printf("%v", err)
			return []tabulaeModels.Email{}, apiModels.User{}, apiModels.Billing{}, []tabulaeModels.File{}, err
		}
	}

	user := apiModels.User{}
	userId := datastore.IDKey("User", emails[0].CreatedBy, nil)
	if err := datastoreClient.Get(c, userId, &user); err != nil {
		log.Printf("%v", err)
		return []tabulaeModels.Email{}, apiModels.User{}, apiModels.Billing{}, []tabulaeModels.File{}, err
	}
	user.Id = emails[0].CreatedBy

	userBillings := []apiModels.Billing{}
	q := datastore.NewQuery("Billing").Filter("CreatedBy =", user.Id).Limit(1)
	_, err := datastoreClient.GetAll(c, q, &userBillings)
	if err != nil {
		log.Printf("%v", err)
		return []tabulaeModels.Email{}, apiModels.User{}, apiModels.Billing{}, []tabulaeModels.File{}, err
	}

	if len(userBillings) == 0 {
		err := errors.New("Missing user billing")
		return emails, user, apiModels.Billing{}, files, err
	}

	return emails, user, userBillings[0], files, nil
}

func getAttachment(c context.Context, file tabulaeModels.File) ([]byte, string, string, error) {
	client, err := storage.NewClient(c)
	if err != nil {
		return nil, "", "", err
	}
	defer client.Close()

	clientBucket := client.Bucket("tabulae-email-attachment")
	rc, err := clientBucket.Object(file.FileName).NewReader(c)
	if err != nil {
		return nil, "", "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, "", "", err
	}

	return data, rc.ContentType(), file.OriginalName, nil
}

func getAttachments(c context.Context, files []tabulaeModels.File) ([][]byte, []string, []string, error) {
	if len(files) == 0 {
		return [][]byte{}, []string{}, []string{}, nil
	}

	bytesArray := [][]byte{}
	attachmentTypes := []string{}
	fileNames := []string{}
	for i := 0; i < len(files); i++ {
		currentBytes, attachmentType, fileName, err := getAttachment(c, files[i])
		if err == nil {
			bytesArray = append(bytesArray, currentBytes)
			attachmentTypes = append(attachmentTypes, attachmentType)
			fileNames = append(fileNames, fileName)
		} else {
			log.Printf("%v", err)
		}
	}

	return bytesArray, attachmentTypes, fileNames, nil
}

func subscribe() {
	c := context.Background()
	err := subscription.Receive(c, func(c context.Context, msg *pubsub.Message) {
		var ids map[string][]int64
		if err := json.Unmarshal(msg.Data, &ids); err != nil {
			log.Printf("%v", err)
			log.Printf("error decoding message data: %#v", msg)
			msg.Ack()
			return
		}

		// Get emails, and details surrounding the emails
		allEmails, user, userBilling, files, err := getEmails(c, ids["EmailIds"])
		if err != nil {
			log.Printf("error getting emails: %v", err)
			msg.Ack()
			return
		}

		// Get the actual attachment files from Google storage
		bytesArray, attachmentType, fileNames, err := getAttachments(c, files)
		if err != nil {
			log.Printf("error getting attachments: %v", err)
			msg.Ack()
			return
		}

		// Check for duplicates in email Ids. In-case a "send" is clicked two
		// for one particular email.
		newEmails := []tabulaeModels.Email{}

		// These delay parameters are for the delay between emails
		betweenDelay := 60
		sendGridKey := getSendGridKeyForUser(userBilling)
		for i := 0; i < len(allEmails); i++ {
			if allEmails[i].Method == "sendgrid" {
				// Select delay to be used using Sendgrid
				delayAmount := int(float64(i) / float64(200))
				sendGridDelay := delayAmount * betweenDelay

				emailWithId, _, err := sendSendGridEmail(c, allEmails[i], files, user, bytesArray, attachmentType, fileNames, sendGridKey, sendGridDelay)
				if err != nil {
					log.Printf("error emailId: %v", emailWithId.Id)
					log.Printf("error sending sendgrid email: %v", err)
					continue
				}
				newEmails = append(newEmails, emailWithId)
				// } else if allEmails[i].Method == "gmail" {
			} else if allEmails[i].Method == "outlook" {
				emailWithId, _, err := sendOutlookEmail(c, allEmails[i], files, user, bytesArray, attachmentType, fileNames)
				if err != nil {
					log.Printf("error emailId: %v", emailWithId.Id)
					log.Printf("error sending outlook email: %v", err)
					continue
				}
				newEmails = append(newEmails, emailWithId)
			} else if allEmails[i].Method == "smtp" {
				emailWithId, _, err := sendSMTPEmail(c, allEmails[i], files, user, bytesArray, attachmentType, fileNames)
				if err != nil {
					log.Printf("error emailId: %v", emailWithId.Id)
					log.Printf("error sending smtp email: %v", err)
					continue
				}
				newEmails = append(newEmails, emailWithId)
			} else if allEmails[i].Method == "gmail" {
				emailWithId, _, err := sendGmailEmail(c, allEmails[i], files, user, bytesArray, attachmentType, fileNames)
				if err != nil {
					log.Printf("error emailId: %v", emailWithId.Id)
					log.Printf("error sending gmail email: %v", err)
					continue
				}
				newEmails = append(newEmails, emailWithId)
			}
		}

		// Send message to updates-service that the data has been changed/updated
		updates := []updateService.EmailSendUpdate{}
		for i := 0; i < len(newEmails); i++ {
			update := updateService.EmailSendUpdate{}
			update.EmailId = newEmails[i].Id
			update.Method = newEmails[i].Method
			update.Delievered = true

			if update.Method == "gmail" {
				update.SendId = newEmails[i].GmailId
				update.ThreadId = newEmails[i].GmailThreadId
			} else if update.Method == "sendgrid" {
				update.SendId = newEmails[i].SendGridId
			} else if update.Method == "outlook" {
				// no updates from Outlook
			} else if update.Method == "smtp" {
				// no updates from SMTP
			}

			updates = append(updates, update)
		}

		if len(updates) > 0 {
			err = sendChangesToUpdateService(c, updates)
			if err != nil {
				log.Printf("error updates: %v", updates)
				log.Printf("error sending emails to update service: %v", err)
			}
		}

		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	projectID := "newsai-1166"
	c := context.Background()

	pubsubClient, err := pubsub.NewClient(c, projectID)
	if err != nil {
		log.Fatal(err)
		return
	}

	topic := pubsubClient.Topic("tabulae-emails-service")
	subscription, _ = pubsubClient.CreateSubscription(c, "appengine-flex-service-1", pubsub.SubscriptionConfig{Topic: topic})

	datastoreClient, err = datastore.NewClient(c, projectID)
	if err != nil {
		log.Fatal(err)
		return
	}

	go subscribe()
	http.HandleFunc("/", handle)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "E M A I L S")
}
