package sync

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/pubsub"
)

var (
	PubsubClient        *pubsub.Client
	InfluencerTopicID   = "influencer"
	ContactsTopicID     = "datastore-sync-contacts-functions"
	PublicationsTopicID = "datastore-sync-publications-functions"
	ListsTopicID        = "datastore-sync-lists-functions"
	TwitterTopicID      = "process-twitter-feed"
	projectID           = "newsai-1166"
)

func configurePubsub(r *http.Request) (*pubsub.Client, error) {
	if PubsubClient != nil {
		return PubsubClient, nil
	}
	c := appengine.NewContext(r)
	PubsubClient, err := pubsub.NewClient(c, projectID)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	// Create the topic for influencers if it doesn't exist.
	if exists, err := PubsubClient.Topic(InfluencerTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(c, InfluencerTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for contacts if it doesn't exist.
	if exists, err := PubsubClient.Topic(ContactsTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(c, ContactsTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for publications if it doesn't exist.
	if exists, err := PubsubClient.Topic(PublicationsTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(c, PublicationsTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for publications if it doesn't exist.
	if exists, err := PubsubClient.Topic(ListsTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(c, ListsTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for twitter if it doesn't exist.
	if exists, err := PubsubClient.Topic(TwitterTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(c, TwitterTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	return PubsubClient, nil
}
