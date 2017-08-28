package main

import (
	"encoding/base64"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/news-ai/tabulae/models"

	"golang.org/x/net/context"

	apiModels "github.com/news-ai/api/models"
	tabulaeModels "github.com/news-ai/tabulae/models"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func getSendGridKeyForUser(userBilling apiModels.Billing) string {
	if userBilling.IsOnTrial {
		return os.Getenv("SENDGRID_INTERNAL_API_KEY")
	}

	return os.Getenv("SENDGRID_API_KEY")
}

func sendSendGridEmail(c context.Context, email tabulaeModels.Email, files []tabulaeModels.File, user apiModels.User, bytesArray [][]byte, attachmentType []string, fileNames []string, sendGridKey string, sendGridDelay int) (tabulaeModels.Email, interface{}, error) {
	email.Method = "sendgrid"
	email.IsSent = true

	// Test if the email we are sending with is in the user's SendGridFrom or is their Email
	if email.FromEmail != "" {
		userEmailValid := false
		if user.Email == email.FromEmail {
			userEmailValid = true
		}

		for i := 0; i < len(user.Emails); i++ {
			if user.Emails[i] == email.FromEmail {
				userEmailValid = true
			}
		}

		// If this is if the email added is not valid in SendGridFrom
		if !userEmailValid {
			return email, nil, errors.New("The email requested is not confirmed by the user yet")
		}
	}

	// Check to see if there is no sendat date or if date is in the past
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		emailSent, emailId, err := sendEmailAttachment(c, email, user, files, bytesArray, attachmentType, fileNames, sendGridKey, sendGridDelay)
		if err != nil {
			log.Printf("%v", err)
			return tabulaeModels.Email{}, nil, err
		}

		email.IsSent = emailSent
		email.Delievered = emailSent
		email.SendGridId = emailId
		return email, nil, nil
	}

	return tabulaeModels.Email{}, nil, nil
}

// Send an email confirmation to a new user
func sendEmailAttachment(c context.Context, email models.Email, user apiModels.User, files []models.File, bytesArray [][]byte, attachmentType []string, fileNames []string, sendGridKey string, sendGridDelay int) (bool, string, error) {
	userFullName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	emailFullName := strings.Join([]string{email.FirstName, email.LastName}, " ")

	from := mail.NewEmail(userFullName, user.Email)

	if user.EmailAlias != "" {
		from = mail.NewEmail(userFullName, user.EmailAlias)
	}

	if email.FromEmail != "" {
		from = mail.NewEmail(userFullName, email.FromEmail)
	}

	to := mail.NewEmail(emailFullName, email.To)
	content := mail.NewContent("text/html", email.Body)

	m := mail.NewV3Mail()

	// Set from
	m.SetFrom(from)
	m.Content = []*mail.Content{
		content,
	}

	// Adding a personalization for the email
	p := mail.NewPersonalization()

	if email.Subject == "" {
		p.Subject = "(no subject)"
	} else {
		p.Subject = email.Subject
	}

	// Adding who we are sending the email to
	tos := []*mail.Email{
		to,
	}

	p.AddTos(tos...)

	ccs := []*mail.Email{}
	for i := 0; i < len(email.CC); i++ {
		cc := mail.NewEmail("", email.CC[i])
		ccs = append(ccs, cc)
	}

	if len(ccs) > 0 {
		p.AddCCs(ccs...)
	}

	bccs := []*mail.Email{}
	for i := 0; i < len(email.BCC); i++ {
		bcc := mail.NewEmail("", email.BCC[i])
		bccs = append(bccs, bcc)
	}

	if len(bccs) > 0 {
		p.AddBCCs(bccs...)
	}

	// Add personalization
	m.AddPersonalizations(p)

	// Add attachments
	if len(files) > 0 {
		for i := 0; i < len(bytesArray); i++ {
			a := mail.NewAttachment()
			str := base64.StdEncoding.EncodeToString(bytesArray[i])

			a.SetContent(str)
			a.SetType(attachmentType[i])
			a.SetFilename(fileNames[i])
			a.SetDisposition("attachment")
			m.AddAttachment(a)
		}
	}

	if sendGridDelay > 0 {
		timeSend := time.Now()
		timeSend = timeSend.Add(time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(sendGridDelay))

		timeInt := int(timeSend.Unix())
		m.SetSendAt(timeInt)
	}

	request := sendgrid.GetRequest(sendGridKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)

	// Send the actual mail here
	response, err := sendgrid.API(request)
	if err != nil {
		log.Printf("error: %v", err)
		return false, "", err
	}

	emailId := ""
	if len(response.Headers["X-Message-Id"]) > 0 {
		emailId = response.Headers["X-Message-Id"][0]
	}

	return true, emailId, nil
}
