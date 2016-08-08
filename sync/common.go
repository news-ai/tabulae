package sync

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/cloud/pubsub"
)

var (
	PubsubClient  *pubsub.Client
	PubsubTopicID = "influencer"
	projectID     = "newsai-1166"
)

func configurePubsub(r *http.Request) (*pubsub.Client, error) {
	if PubsubClient != nil {
		return PubsubClient, nil
	}
	c := appengine.NewContext(r)
	PubsubClient, err := pubsub.NewClient(c, projectID)
	if err != nil {
		return nil, err
	}

	// Create the topic if it doesn't exist.
	if exists, err := PubsubClient.Topic(PubsubTopicID).Exists(c); err != nil {
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(c, PubsubTopicID); err != nil {
			return nil, err
		}
	}
	return PubsubClient, nil
}
