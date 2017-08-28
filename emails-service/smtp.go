package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"

	"golang.org/x/net/context"

	apiModels "github.com/news-ai/api/models"
	tabulaeModels "github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/emails"
)

type SMTPEmailResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

func getEmailSetting(c context.Context, id int64) (tabulaeModels.EmailSetting, error) {
	emailSetting := tabulaeModels.EmailSetting{}
	emailSettingId := datastore.IDKey("EmailSetting", id, nil)
	if err := datastoreClient.Get(c, emailSettingId, &emailSetting); err != nil {
		log.Printf("%v", err)
		return tabulaeModels.EmailSetting{}, err
	}

	emailSetting.Id = id
	return emailSetting, nil
}

func sendSMTPEmail(c context.Context, email tabulaeModels.Email, files []tabulaeModels.File, user apiModels.User, bytesArray [][]byte, attachmentType []string, fileNames []string) (tabulaeModels.Email, interface{}, error) {
	email.IsSent = true

	// Check to see if there is no sendat date or if date is in the past
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		emailBody, err := emails.GenerateEmail(user, email, files, bytesArray, attachmentType, fileNames)
		if err != nil {
			log.Printf("%v", err)
			return tabulaeModels.Email{}, nil, err
		}

		emailSetting, err := getEmailSetting(c, user.EmailSetting)
		if err != nil {
			log.Printf("%v", err)
			return tabulaeModels.Email{}, nil, err
		}

		SMTPPassword := string(user.SMTPPassword[:])

		getUrl := "https://tabulae-smtp.newsai.org/send"

		sendEmailRequest := tabulaeModels.SMTPEmailSettings{}
		sendEmailRequest.Servername = emailSetting.SMTPServer + ":" + strconv.Itoa(emailSetting.SMTPPortSSL)
		sendEmailRequest.EmailUser = user.SMTPUsername
		sendEmailRequest.EmailPassword = SMTPPassword
		sendEmailRequest.To = email.To
		sendEmailRequest.Subject = email.Subject
		sendEmailRequest.Body = emailBody

		SendEmailRequest, err := json.Marshal(sendEmailRequest)
		if err != nil {
			log.Printf("%v", err)
			return tabulaeModels.Email{}, nil, err
		}
		log.Printf("%v", string(SendEmailRequest))
		sendEmailQuery := bytes.NewReader(SendEmailRequest)

		req, _ := http.NewRequest("POST", getUrl, sendEmailQuery)

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%v", err)
			return tabulaeModels.Email{}, nil, err
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var verifyResponse SMTPEmailResponse
		err = decoder.Decode(&verifyResponse)
		if err != nil {
			log.Printf("%v", err)
			return tabulaeModels.Email{}, nil, err
		}

		log.Printf("%v", verifyResponse)

		if verifyResponse.Status {
			email.Delievered = true
			return email, nil, nil
		}

		return tabulaeModels.Email{}, nil, errors.New(verifyResponse.Error)
	}

	return tabulaeModels.Email{}, nil, nil
}
