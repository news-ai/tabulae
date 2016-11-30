package emails

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/news-ai/tabulae/models"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type CampaignMonitorAddSubscriber struct {
	EmailAddress string `json:"EmailAddress"`
	Name         string `json:"Name"`
	CustomFields []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"CustomFields"`
	Resubscribe                            bool `json:"Resubscribe"`
	RestartSubscriptionBasedAutoresponders bool `json:"RestartSubscriptionBasedAutoresponders"`
}

func AddUserToTabulaeTrialList(c context.Context, user models.User) error {
	apiKey := os.Getenv("CAMPAIGNMONITOR_API_KEY")
	trialListId := "7dc0d29f2d1ba1c0bda15e74f57599bc"

	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)

	newSubscriber := CampaignMonitorAddSubscriber{}
	newSubscriber.EmailAddress = user.Email
	newSubscriber.Name = user.FirstName + " " + user.LastName
	newSubscriber.Resubscribe = true
	newSubscriber.RestartSubscriptionBasedAutoresponders = false

	NewSubscriber, err := json.Marshal(newSubscriber)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}
	newSubscriberJson := bytes.NewReader(NewSubscriber)

	postUrl := "https://api.createsend.com/api/v3.1/subscribers/" + trialListId + ".json"
	req, _ := http.NewRequest("POST", postUrl, newSubscriberJson)
	req.SetBasicAuth(apiKey, "x")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	if resp.StatusCode == 201 {
		return nil
	}

	return errors.New("Error happened when sending email")
}
