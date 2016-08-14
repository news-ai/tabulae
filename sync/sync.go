package sync

import (
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/pubsub"
)

func LinkedInSync(r *http.Request, contactLinkedIn string, contactId int64, justCreated bool) error {
	c := appengine.NewContext(r)
	PubsubClient, err := configurePubsub(r)
	if err != nil {
		return err
	}

	// Create an map with linkedinUrl and Id of the corresponding contact
	data := map[string]string{
		"Id":          strconv.FormatInt(contactId, 10),
		"linkedinUrl": contactLinkedIn,
		"justCreated": strconv.FormatBool(justCreated),
	}

	log.Infof(c, "%v", data)

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
