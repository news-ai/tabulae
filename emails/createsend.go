package emails

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
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

type CampaignMonitorResetEmail struct {
	To   []string `json:"To"`
	Data struct {
		RESET_CODE string `json:"RESET_CODE"`
	} `json:"Data"`
	AddRecipientsToList bool `json:"AddRecipientsToList"`
}

type CampaignMonitorConfirmationEmail struct {
	To   []string `json:"To"`
	Data struct {
		CONFIRMATION_CODE string `json:"CONFIRMATION_CODE"`
	} `json:"Data"`
	AddRecipientsToList bool `json:"AddRecipientsToList"`
}

type CampaignMonitorPremiumEmail struct {
	To   []string `json:"To"`
	Data struct {
		PLAN       string `json:"PLAN"`
		DURATION   string `json:"DURATION"`
		BILLDATE   string `json:"BILLDATE"`
		BILLAMOUNT string `json:"BILLAMOUNT"`
	} `json:"Data"`
	AddRecipientsToList bool `json:"AddRecipientsToList"`
}

func ConfirmUserAccount(c context.Context, user models.User, confirmationCode string) error {
	apiKey := os.Getenv("CAMPAIGNMONITOR_API_KEY")
	confirmationEmailId := "a609aac8-cde6-4830-92ba-215ee48c4195"

	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)

	confirmationEmail := CampaignMonitorConfirmationEmail{}

	userEmail := user.FirstName + " " + user.LastName + " <" + user.Email + " >"
	confirmationEmail.To = append(confirmationEmail.To, userEmail)
	confirmationEmail.AddRecipientsToList = false

	t := &url.URL{Path: confirmationCode}
	encodedConfirmationCode := t.String()
	confirmationEmail.Data.CONFIRMATION_CODE = encodedConfirmationCode

	ConfirmationEmail, err := json.Marshal(confirmationEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}
	confirmationEmailJson := bytes.NewReader(ConfirmationEmail)

	postUrl := "https://api.createsend.com/api/v3.1/transactional/smartEmail/" + confirmationEmailId + "/send"

	req, _ := http.NewRequest("POST", postUrl, confirmationEmailJson)
	req.SetBasicAuth(apiKey, "x")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	if resp.StatusCode == 201 || resp.StatusCode == 202 || resp.StatusCode == 200 {
		return nil
	}

	return errors.New("Error happened when sending email")
}

func ResetUserPassword(c context.Context, user models.User, resetPasswordCode string) error {
	apiKey := os.Getenv("CAMPAIGNMONITOR_API_KEY")
	resetEmailId := "b85b1152-5665-46ff-ada8-a5720b730a51"

	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)

	resetEmail := CampaignMonitorResetEmail{}

	userEmail := user.FirstName + " " + user.LastName + " <" + user.Email + " >"
	resetEmail.To = append(resetEmail.To, userEmail)
	resetEmail.AddRecipientsToList = false

	t := &url.URL{Path: resetPasswordCode}
	encodedResetCode := t.String()
	resetEmail.Data.RESET_CODE = encodedResetCode

	ResetEmail, err := json.Marshal(resetEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}
	resetEmailJson := bytes.NewReader(ResetEmail)

	postUrl := "https://api.createsend.com/api/v3.1/transactional/smartEmail/" + resetEmailId + "/send"

	req, _ := http.NewRequest("POST", postUrl, resetEmailJson)
	req.SetBasicAuth(apiKey, "x")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	if resp.StatusCode == 201 || resp.StatusCode == 202 || resp.StatusCode == 200 {
		return nil
	}

	return errors.New("Error happened when sending email")
}

func AddUserToTabulaePremiumList(c context.Context, user models.User, plan string, duration string, billDate string, billAmount string, paidAmount string) error {
	apiKey := os.Getenv("CAMPAIGNMONITOR_API_KEY")
	premiumEmailId := "62b31c10-4e4d-4d9f-8442-8834427b2040"

	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)

	premiumEmail := CampaignMonitorPremiumEmail{}

	userEmail := user.FirstName + " " + user.LastName + " <" + user.Email + " >"
	premiumEmail.To = append(premiumEmail.To, userEmail)
	premiumEmail.AddRecipientsToList = true

	premiumEmail.Data.PLAN = plan
	premiumEmail.Data.DURATION = duration
	premiumEmail.Data.BILLDATE = billDate
	premiumEmail.Data.BILLAMOUNT = billAmount

	PremiumEmail, err := json.Marshal(premiumEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}
	premiumEmailJson := bytes.NewReader(PremiumEmail)

	postUrl := "https://api.createsend.com/api/v3.1/transactional/smartEmail/" + premiumEmailId + "/send"

	req, _ := http.NewRequest("POST", postUrl, premiumEmailJson)
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
