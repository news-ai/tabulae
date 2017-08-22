package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "strconv"

	"golang.org/x/net/context"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"

	tabulaeModels "github.com/news-ai/tabulae/models"
)

var (
	pubsubClient    *pubsub.Client
	subscription    *pubsub.Subscription
	datastoreClient *datastore.Client
)

func getEmails(c context.Context, ids []int64) ([]tabulaeModels.Email, error) {
	var keys []*datastore.Key
	for i := 0; i < len(ids); i++ {
		emailId := datastore.IDKey("Email", ids[i], nil)
		keys = append(keys, emailId)
	}

	emails := make([]tabulaeModels.Email, len(keys))
	if err := datastoreClient.GetMulti(c, keys, emails); err != nil {
		log.Printf("%v", err)
		return []tabulaeModels.Email{}, err
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
