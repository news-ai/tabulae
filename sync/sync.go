package sync

import (
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/pubsub"
)

func sync(r *http.Request, data map[string]string, topicName string) error {
	c := appengine.NewContext(r)
	PubsubClient, err := configurePubsub(r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	topic := PubsubClient.Topic(topicName)
	_, err = topic.Publish(c, &pubsub.Message{Data: jsonData})
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	return nil
}

func TwitterSync(r *http.Request, socialField string, twitterUser string, contactId int64) error {
	// Create an map with twitter username and parent Id of the corresponding contact
	data := map[string]string{
		"contactId": strconv.FormatInt(contactId, 10),
		"username":  twitterUser,
	}

	return sync(r, data, TwitterTopicID)
}

func SocialSync(r *http.Request, socialField string, url string, contactId int64, justCreated bool) error {
	// Create an map with linkedinUrl and Id of the corresponding contact
	data := map[string]string{
		"Id":          strconv.FormatInt(contactId, 10),
		socialField:   url,
		"justCreated": strconv.FormatBool(justCreated),
	}

	return sync(r, data, InfluencerTopicID)
}

func ResourceSync(r *http.Request, resourceId int64, resource string) error {
	data := map[string]string{
		"Id": strconv.FormatInt(resourceId, 10),
	}

	topicName := ""

	if resource == "Contact" {
		topicName = ContactsTopicID
	} else if resource == "Publication" {
		topicName = PublicationsTopicID
	} else if resource == "List" {
		return nil
	}

	return sync(r, data, topicName)
}
