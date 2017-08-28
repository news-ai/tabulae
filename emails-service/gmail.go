package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"golang.org/x/net/context"

	apiModels "github.com/news-ai/api/models"
	Gmail "github.com/news-ai/go-gmail"
	"github.com/news-ai/tabulae/models"
	tabulaeModels "github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/google"
)

func sendGmailEmail(c context.Context, email tabulaeModels.Email, files []tabulaeModels.File, user apiModels.User, bytesArray [][]byte, attachmentType []string, fileNames []string) (tabulaeModels.Email, interface{}, error) {
	email.Method = "gmail"
	email.IsSent = true

	err := google.ValidateAccessToken(c, user)
	// Refresh access token if err is nil
	if err != nil {
		log.Printf("%v", err)
		user, err = google.RefreshAccessToken(c, user)
		if err != nil {
			log.Printf("%v", err)
			return email, nil, errors.New("Could not refresh user token")
		}
	}

	// Check to see if there is no sendat date or if date is in the past
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		gmailId, gmailThreadId, err := sendGmailEmailAPI(c, user, email, files, bytesArray, attachmentType, fileNames)
		if err != nil {
			log.Printf("%v", err)
			return email, nil, err
		}

		email.GmailId = gmailId
		email.GmailThreadId = gmailThreadId
		email.Delievered = true
		return email, nil, nil
	}

	return tabulaeModels.Email{}, nil, nil
}

func sendGmailEmailAPI(c context.Context, user apiModels.User, email models.Email, files []models.File, bytesArray [][]byte, attachmentType []string, fileNames []string) (string, string, error) {
	userFullName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	emailFullName := strings.Join([]string{email.FirstName, email.LastName}, " ")

	from := userFullName + "<" + user.Email + ">"

	if user.EmailAlias != "" {
		from = userFullName + "<" + user.EmailAlias + ">"
	}

	to := emailFullName + "<" + email.To + ">"

	gmail := Gmail.Gmail{}
	gmail.AccessToken = user.AccessToken

	if len(email.Attachments) > 0 && len(files) > 0 {
		return gmail.SendEmailWithAttachments(c, from, to, email.Subject, email.Body, email, files, bytesArray, attachmentType, fileNames)
	}

	return gmail.SendEmail(c, from, to, email.Subject, email.Body, email)
}
