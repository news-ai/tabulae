package sync

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	// Create an map with RSS feed url and publicationId
	data := map[string]string{
		"url":           url,
		"publicationId": strconv.FormatInt(publicationId, 10),
	}

	return sync(r, data, RSSFeedTopicID)
}

func InstagramSync(r *http.Request, instagramUser string, instagramAccessToken string) error {
	// Create an map with instagram username and instagramAccessToken
	if instagramUser != "" {
		data := map[string]string{
			"username":     instagramUser,
			"access_token": "",
		}

		return sync(r, data, InstagramTopicID)
	}

	return errors.New("Instagram username is not valid")
}

func TwitterSync(r *http.Request, twitterUser string) error {
	// Create an map with twitter username
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

func EmailResourceBulkSync(r *http.Request, emailIds []int64) error {
	tempEmailResourceIds := []string{}
	for i := 0; i < len(emailIds); i++ {
		tempEmailResourceIds = append(tempEmailResourceIds, strconv.FormatInt(emailIds[i], 10))
	}

	topicName := EmailBulkTopicID
	data := map[string]string{
		"EmailId": strings.Join(tempEmailResourceIds, ","),
		"Method":  "create",
	}

	err := sync(r, data, topicName)
	if err != nil {
		return err
	}

	return nil
}

func ListUploadResourceBulkSync(r *http.Request, listId int64, contactIds []int64, publicationIds []int64) error {
	tempContactResourceIds := []string{}
	for i := 0; i < len(contactIds); i++ {
		tempContactResourceIds = append(tempContactResourceIds, strconv.FormatInt(contactIds[i], 10))
	}

	tempPublicationResourceIds := []string{}
	for i := 0; i < len(publicationIds); i++ {
		tempPublicationResourceIds = append(tempPublicationResourceIds, strconv.FormatInt(publicationIds[i], 10))
	}

	topicName := ListUploadTopicID
	data := map[string]string{
		"ListId":        strconv.FormatInt(listId, 10),
		"PublicationId": strings.Join(tempPublicationResourceIds, ","),
		"ContactId":     strings.Join(tempContactResourceIds, ","),
		"Method":        "create",
	}

	err := sync(r, data, topicName)
	if err != nil {
		return err
	}

	return nil
}

func ResourceSync(r *http.Request, resourceId int64, resource string, method string) error {
	data := map[string]string{
		"Id":     strconv.FormatInt(resourceId, 10),
		"Method": method,
	}

	topicName := ""

	if resource == "Contact" {
		topicName = ContactChangeTopicID
	} else if resource == "Publication" {
		topicName = PublicationChangeTopicID
	} else if resource == "List" {
		topicName = ListChangeTopicID
	} else if resource == "User" {
		topicName = UserChangeTopicID
	} else if resource == "Email" {
		topicName = EmailChangeTopicID
	}

	return sync(r, data, topicName)
}
