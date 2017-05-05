package sync

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/pubsub"
)

var (
	PubsubClient             *pubsub.Client
	InfluencerTopicID        = "influencer"
	ListChangeTopicId        = "process-list-change"
	EmailChangeTopicID       = "process-email-change"
	EmailBulkTopicID         = "process-email-change-bulk"
	ContactChangeTopicID     = "process-contact-change"
	UserChangeTopicID        = "process-user-change"
	PublicationChangeTopicID = "process-new-publication-upload"
	TwitterTopicID           = "process-twitter-feed"
	InstagramTopicID         = "process-instagram-feed"
	EnhanceTopicID           = "process-enhance"
	RSSFeedTopicID           = "process-rss-feed"
	ListUploadTopicID        = "process-new-list-upload"
	projectID                = "newsai-1166"
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

	// Create the topic for instagram if it doesn't exist.
	if exists, err := PubsubClient.Topic(EmailBulkTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, EmailBulkTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	return PubsubClient, nil
}
