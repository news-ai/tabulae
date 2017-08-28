package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"golang.org/x/net/context"

	apiModels "github.com/news-ai/api/models"
	Outlook "github.com/news-ai/go-outlook"
	tabulaeModels "github.com/news-ai/tabulae/models"
	"github.com/news-ai/web/outlook"
)

func sendOutlookEmail(c context.Context, email tabulaeModels.Email, files []tabulaeModels.File, user apiModels.User, bytesArray [][]byte, attachmentType []string, fileNames []string) (tabulaeModels.Email, interface{}, error) {
	err := outlook.ValidateAccessToken(c, user)
	// Refresh access token if err is nil
	if err != nil {
		log.Printf("%v", err)
		user, err = outlook.RefreshAccessToken(c, user)
		if err != nil {
			log.Printf("%v", err)
			return email, nil, errors.New("Could not refresh user token")
		}
	}

	email.IsSent = true

	// Check to see if there is no sendat date or if date is in the past
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		err := sendOutlookEmailAPI(c, user, email, files, bytesArray, attachmentType, fileNames)
		if err != nil {
			log.Printf("%v", err)
			return email, nil, err
		}

		email.Delievered = true
		return email, nil, nil
	}

	return tabulaeModels.Email{}, nil, nil
}

func sendOutlookEmailAPI(c context.Context, user apiModels.User, email tabulaeModels.Email, files []tabulaeModels.File, bytesArray [][]byte, attachmentType []string, fileNames []string) error {
	userFullName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	from := userFullName + "<" + user.OutlookEmail + ">"

	outlook := Outlook.Outlook{}
	outlook.AccessToken = user.OutlookAccessToken

	if len(email.Attachments) > 0 && len(files) > 0 {
		return outlook.SendEmailWithAttachments(c, from, email.To, email.Subject, email.Body, email, files, bytesArray, attachmentType, fileNames)
	}

	return outlook.SendEmail(c, from, email.To, email.Subject, email.Body, email)
}
