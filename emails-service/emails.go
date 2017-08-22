package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"

	apiModels "github.com/news-ai/api/models"
	tabulaeModels "github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/emails"
)

var (
	pubsubClient    *pubsub.Client
	subscription    *pubsub.Subscription
	datastoreClient *datastore.Client
)

func sendSendGridEmail(c context.Context, r *http.Request, email tabulaeModels.Email, files []tabulaeModels.File, user apiModels.User, bytesArray [][]byte, attachmentType []string, fileNames []string, sendGridKey string) (tabulaeModels.Email, interface{}, error) {
	email.Method = "sendgrid"
	email.IsSent = true

	// Test if the email we are sending with is in the user's SendGridFrom or is their Email
	if email.FromEmail != "" {
		userEmailValid := false
		if user.Email == email.FromEmail {
			userEmailValid = true
		}

		for i := 0; i < len(user.Emails); i++ {
			if user.Emails[i] == email.FromEmail {
				userEmailValid = true
			}
		}

		// If this is if the email added is not valid in SendGridFrom
		if !userEmailValid {
			return email, nil, errors.New("The email requested is not confirmed by the user yet")
		}
	}

	// Check to see if there is no sendat date or if date is in the past
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		emailSent, emailId, err := emails.SendEmailAttachment(r, email, user, files, bytesArray, attachmentType, fileNames, sendGridKey)
		email.IsSent = true
		email.Delievered = true
		email.SendGridId = emailId
	}
}

func getEmails(c context.Context, ids []int64) ([]tabulaeModels.Email, error) {
	var keys []*datastore.Key
	for i := 0; i < len(ids); i++ {
		emailId := datastore.IDKey("Email", ids[i], nil)
		log.Println(emailId.Encode())
		keys = append(keys, emailId)
	}

	emails := make([]tabulaeModels.Email, len(keys))
	if err := datastoreClient.GetMulti(c, keys, emails); err != nil {
		log.Printf("%v", err)
		return []tabulaeModels.Email{}, err
	}

	// Remove emails that have already been delivered
	// Downward loop fixes problems that you have when deleting
	// elements in an array while looping through them.
	for i := len(emails) - 1; i >= 0; i-- {
		if emails[i].Delievered {
			emails = append(emails[:i], emails[i+1:]...)
		}
	}

	return emails, nil
}

func subscribe() {
	c := context.Background()
	err := subscription.Receive(c, func(c context.Context, msg *pubsub.Message) {
		var ids []int64
		if err := json.Unmarshal(msg.Data, &ids); err != nil {
			log.Printf("%v", err)
			log.Printf("could not decode message data: %#v", msg)
			msg.Ack()
			return
		}

		emails, err := getEmails(c, ids)
		if err != nil {
			log.Printf("%v", err)
			msg.Ack()
			return
		}

		log.Printf("%v", emails)

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
