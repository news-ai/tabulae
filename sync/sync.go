package sync

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/cloud/pubsub"
)

type LinkedInData struct {
	Current []struct {
		Date     string `json:"date"`
		Position string `json:"position"`
		Employer string `json:"employer"`
	} `json:"current"`
	Past []struct {
		Date     string `json:"date"`
		Position string `json:"position"`
		Employer string `json:"employer"`
	} `json:"past"`
}

func LinkedInSync(r *http.Request, contactLinkedIn string) error {
	c := appengine.NewContext(r)
	PubsubClient, err := configurePubsub(r)
	if err != nil {
		return err
	}

	topic := PubsubClient.Topic(PubsubTopicID)
	_, err = topic.Publish(c, &pubsub.Message{Data: []byte(contactLinkedIn)})
	if err != nil {
		return err
	}

	return nil
}
