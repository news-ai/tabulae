package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"cloud.google.com/go/pubsub"
)

var (
	PubsubClient *pubsub.Client
	subscription *pubsub.Subscription
)

func subscribe() {
	c := context.Background()
	err := subscription.Receive(c, func(c context.Context, msg *pubsub.Message) {
		var ids []int64
		if err := json.Unmarshal(msg.Data, &ids); err != nil {
			log.Printf("could not decode message data: %#v", msg)
			msg.Ack()
			return
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	c := context.Background()

	PubsubClient, err := pubsub.NewClient(c, "newsai-1166")
	topic := PubsubClient.Topic("tabulae-emails-service")
	subscription, _ = PubsubClient.CreateSubscription(c, "appengine-flex-service-1", pubsub.SubscriptionConfig{Topic: topic})
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
