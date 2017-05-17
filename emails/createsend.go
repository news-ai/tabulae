package emails

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
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

type CampaignMonitorAddUserEmail struct {
	To   []string `json:"To"`
	Data struct {
		ADD_EMAIL_CODE string `json:"ADD_EMAIL_CODE"`
	} `json:"Data"`
	AddRecipientsToList bool `json:"AddRecipientsToList"`
}

type CampaignMonitorInviteUserEmail struct {
	To   []string `json:"To"`
	Data struct {
		INVITE_EMAIL_CODE     string `json:"INVITE_EMAIL_CODE"`
		NEWUSER_EMAIL         string `json:"NEWUSER_EMAIL"`
		PERSONAL_MESSAGE      string `json:"PERSONAL_MESSAGE"`
		CURRENTUSER_FULL_NAME string `json:"CURRENTUSER_FULL_NAME"`
		CURRENTUSER_EMAIL     string `json:"CURRENTUSER_EMAIL"`
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

func AddUserToTabulaePremiumList(c context.Context, user models.User, plan, duration, billDate, billAmount, paidAmount string) error {
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

func AddEmailToUser(c context.Context, user models.User, userToEmail, userEmailCode string) error {
	apiKey := os.Getenv("CAMPAIGNMONITOR_API_KEY")
	addEmailCodeId := "3cf262ae-51aa-4735-a163-c570a8e861b3"

	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)

	addEmail := CampaignMonitorAddUserEmail{}

	userEmail := user.FirstName + " " + user.LastName + " <" + userToEmail + " >"
	addEmail.To = append(addEmail.To, userEmail)
	addEmail.AddRecipientsToList = false

	t := &url.URL{Path: userEmailCode}
	encodedUserEmailCode := t.String()
	addEmail.Data.ADD_EMAIL_CODE = encodedUserEmailCode

	AddUserEmail, err := json.Marshal(addEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}
	addUserEmailJson := bytes.NewReader(AddUserEmail)

	postUrl := "https://api.createsend.com/api/v3.1/transactional/smartEmail/" + addEmailCodeId + "/send"

	req, _ := http.NewRequest("POST", postUrl, addUserEmailJson)
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

func InviteUser(c context.Context, currentUser models.User, userEmail, userReferralCode, personalMessage string) error {
	apiKey := os.Getenv("CAMPAIGNMONITOR_API_KEY")
	inviteUserCodeId := ""

	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)

	inviteEmail := CampaignMonitorInviteUserEmail{}
	inviteEmail.To = append(inviteEmail.To, userEmail)
	inviteEmail.AddRecipientsToList = false

	t := &url.URL{Path: userReferralCode}
	encodedUserInviteCode := t.String()
	inviteEmail.Data.INVITE_EMAIL_CODE = encodedUserInviteCode

	inviteEmail.Data.CURRENTUSER_EMAIL = currentUser.Email
	inviteEmail.Data.CURRENTUSER_FULL_NAME = strings.Join([]string{currentUser.FirstName, currentUser.LastName}, " ")
	inviteEmail.Data.PERSONAL_MESSAGE = personalMessage
	inviteEmail.Data.NEWUSER_EMAIL = userEmail

	InviteUserEmail, err := json.Marshal(inviteEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}
	inviteUserEmailJson := bytes.NewReader(InviteUserEmail)

	postUrl := "https://api.createsend.com/api/v3.1/transactional/smartEmail/" + inviteUserCodeId + "/send"

	req, _ := http.NewRequest("POST", postUrl, inviteUserEmailJson)
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
