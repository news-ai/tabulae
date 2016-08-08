package sync

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func LinkedInSync(r *http.Request, contactLinkedIn string, contactId int64) error {
	c := appengine.NewContext(r)
	PubsubClient, err := configurePubsub(r)
	if err != nil {
		return err
	}

	// Create an map with linkedinUrl and Id of the corresponding contact
	data := map[string]string{
		"Id":          strconv.FormatInt(contactId, 10),
		"linkedinUrl": contactLinkedIn,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	topic := PubsubClient.Topic(PubsubTopicID)
	_, err = topic.Publish(c, &pubsub.Message{Data: jsonData})
	if err != nil {
		return err
	}

	return nil
}
