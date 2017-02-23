package sync

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/pubsub"
)

var (
	PubsubClient         *pubsub.Client
	InfluencerTopicID    = "influencer"
	ContactsTopicID      = "datastore-sync-contacts-functions"
	PublicationsTopicID  = "datastore-sync-publications-functions"
	ListsTopicID         = "datastore-sync-lists-functions"
	ListChangeTopicId    = "process-list-change"
	ContactChangeTopicID = "process-contact-change"
	UsersTopicID         = "datastore-sync-users-functions"
	TwitterTopicID       = "process-twitter-feed"
	InstagramTopicID     = "process-instagram-feed"
	EnhanceTopicID       = "process-enhance"
	RSSFeedTopicID       = "process-rss-feed"
	ListUploadTopicID    = "process-new-list-upload"
	projectID            = "newsai-1166"
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
		if _, err := PubsubClient.CreateTopic(c, InfluencerTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for contacts if it doesn't exist.
	if exists, err := PubsubClient.Topic(ContactsTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, ContactsTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for publications if it doesn't exist.
	if exists, err := PubsubClient.Topic(PublicationsTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, PublicationsTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for publications if it doesn't exist.
	if exists, err := PubsubClient.Topic(ListsTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, ListsTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for twitter if it doesn't exist.
	if exists, err := PubsubClient.Topic(TwitterTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, TwitterTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for rss if it doesn't exist.
	if exists, err := PubsubClient.Topic(RSSFeedTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, RSSFeedTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for instagram if it doesn't exist.
	if exists, err := PubsubClient.Topic(InstagramTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, InstagramTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for instagram if it doesn't exist.
	if exists, err := PubsubClient.Topic(EnhanceTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, EnhanceTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	// Create the topic for instagram if it doesn't exist.
	if exists, err := PubsubClient.Topic(ListChangeTopicId).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, ListChangeTopicId); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	return PubsubClient, nil
}
