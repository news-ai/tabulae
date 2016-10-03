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

func NewRSSFeedSync(r *http.Request, url string, publicationId int64) error {
	// Create an map with twitter username and parent Id of the corresponding contact
	data := map[string]string{
		"url":           url,
		"publicationId": strconv.FormatInt(publicationId, 10),
	}

	return sync(r, data, RSSFeedTopicID)
}

func TwitterSync(r *http.Request, twitterUser string) error {
	// Create an map with twitter username and parent Id of the corresponding contact
	data := map[string]string{
		"username": twitterUser,
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

func ResourceSync(r *http.Request, resourceId int64, resource string, method string) error {
	data := map[string]string{
		"Id":     strconv.FormatInt(resourceId, 10),
		"Method": method,
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
