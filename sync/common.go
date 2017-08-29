package sync

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"cloud.google.com/go/pubsub"
)

var (
	PubsubClient             *pubsub.Client
	EmailServiceTopicID      = "tabulae-emails-service"
	InfluencerTopicID        = "influencer"
	ListChangeTopicID        = "process-list-change"
	EmailChangeTopicID       = "process-email-change"
	EmailBulkTopicID         = "process-email-change-bulk"
	UserBulkTopicID          = "process-user-change-bulk"
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

	if exists, err := PubsubClient.Topic(InfluencerTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, InfluencerTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(TwitterTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, TwitterTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(RSSFeedTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, RSSFeedTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(InstagramTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, InstagramTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(EnhanceTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, EnhanceTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(ListChangeTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, ListChangeTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(EmailBulkTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, EmailBulkTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(UserBulkTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, UserBulkTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	if exists, err := PubsubClient.Topic(EmailServiceTopicID).Exists(c); err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	} else if !exists {
		if _, err := PubsubClient.CreateTopic(c, EmailServiceTopicID); err != nil {
			log.Errorf(c, "%v", err)
			return nil, err
		}
	}

	return PubsubClient, nil
}
