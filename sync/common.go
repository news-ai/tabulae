package sync

import (
	"google.golang.org/cloud/pubsub"

	"golang.org/x/net/context"
)

var (
	PubsubClient  *pubsub.Client = nil
	PubsubTopicID                = "influencer"
	projectID                    = "newsai-1166"
)

func configurePubsub() error {
	ctx := context.Background()
	PubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	// Create the topic if it doesn't exist.
	if exists, err := PubsubClient.Topic(PubsubTopicID).Exists(ctx); err != nil {
		return err
	} else if !exists {
		if _, err := PubsubClient.NewTopic(ctx, PubsubTopicID); err != nil {
			return err
		}
	}
	return nil
}
