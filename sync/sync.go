package sync

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"cloud.google.com/go/pubsub"
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
	defer topic.Stop()

	var results []*pubsub.PublishResult
	res := topic.Publish(c, &pubsub.Message{Data: jsonData})
	results = append(results, res)
	for _, result := range results {
		id, err := result.Get(c)
		if err != nil {
			log.Infof(c, "%v", err)
			continue
		}
		log.Infof(c, "Published a message with a message ID: %s\n", id)
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

func SendEmailsToEmailService(r *http.Request, emailIds []int64) error {
	if len(emailIds) == 0 {
		return nil
	}

	c := appengine.NewContext(r)
	topicName := EmailServiceTopicID
	data := map[string][]int64{
		"EmailIds": emailIds,
	}

	log.Infof(c, "%v", emailIds)

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
	defer topic.Stop()

	var results []*pubsub.PublishResult
	res := topic.Publish(c, &pubsub.Message{Data: jsonData})
	results = append(results, res)
	for _, result := range results {
		id, err := result.Get(c)
		if err != nil {
			log.Infof(c, "%v", err)
			continue
		}
		log.Infof(c, "Published a message with a message ID: %s\n", id)
	}

	return nil
}

func EmailResourceBulkSync(r *http.Request, emailIds []int64) error {
	if len(emailIds) == 0 {
		return nil
	}

	tempEmailResourceIds := []string{}
	for i := 0; i < len(emailIds); i++ {
		if emailIds[i] != 0 {
			tempEmailResourceIds = append(tempEmailResourceIds, strconv.FormatInt(emailIds[i], 10))
		}
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

func UserResourceBulkSync(r *http.Request, userIds []int64) error {
	if len(userIds) == 0 {
		return nil
	}

	tempUserResourceIds := []string{}
	for i := 0; i < len(userIds); i++ {
		if userIds[i] != 0 {
			tempUserResourceIds = append(tempUserResourceIds, strconv.FormatInt(userIds[i], 10))
		}
	}

	topicName := UserBulkTopicID
	data := map[string]string{
		"UserId": strings.Join(tempUserResourceIds, ","),
		"Method": "create",
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
		if contactIds[i] != 0 {
			tempContactResourceIds = append(tempContactResourceIds, strconv.FormatInt(contactIds[i], 10))
		}
	}

	tempPublicationResourceIds := []string{}
	for i := 0; i < len(publicationIds); i++ {
		if publicationIds[i] != 0 {
			tempPublicationResourceIds = append(tempPublicationResourceIds, strconv.FormatInt(publicationIds[i], 10))
		}
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
