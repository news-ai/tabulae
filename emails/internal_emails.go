package emails

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/news-ai/tabulae/models"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Send an email confirmation to a new user
// Someday convert this to a batch so we can send multiple confirmation emails at once
func SendInternalEmail(r *http.Request, email models.Email, templateId string, subject string, substitution string, confirmationCode string) (bool, string, error) {
	c := appengine.NewContext(r)
	sendgrid.DefaultClient.HTTPClient = urlfetch.Client(c)

	m := mail.NewV3Mail()
	m.SetTemplateID(templateId)

	// Default from from a NewsAI account
	from := mail.NewEmail("Abhi from NewsAI", "abhi@newsai.org")
	m.SetFrom(from)

	// Adding a personalization for the email
	p := mail.NewPersonalization()
	p.Subject = subject

	// Adding who we are sending the email to
	emailFullName := strings.Join([]string{email.FirstName, email.LastName}, " ")
	tos := []*mail.Email{
		mail.NewEmail(emailFullName, email.To),
	}
	p.AddTos(tos...)
	p.SetSubstitution(substitution, confirmationCode)

	// Add personalization
	m.AddPersonalizations(p)

	// Send the email
	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)

	// Send the actual mail here
	response, err := sendgrid.API(request)
	if err != nil {
		log.Errorf(c, "error: %v", err)
		return false, "", err
	}

	emailId := ""
	if len(response.Headers["X-Message-Id"]) > 0 {
		emailId = response.Headers["X-Message-Id"][0]
	}
	return true, emailId, nil
}

func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: confirmationCode}
	encodedConfirmationCode := t.String()
	return SendInternalEmail(r, email, "a64e454c-19d5-4bba-9cef-bd185e7c9b0b", "Thanks for signing up!", "{CONFIRMATION_CODE}", encodedConfirmationCode)
}

func SendResetEmail(r *http.Request, email models.Email, resetPasswordCode string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: resetPasswordCode}
	encodedResetCode := t.String()
	return SendInternalEmail(r, email, "434520df-7773-424a-8e4a-8a6bf1e24441", "Reset your NewsAI password!", "{RESET_CODE}", encodedResetCode)
}

func SendListUploadedEmail(r *http.Request, email models.Email, listId string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: listId}
	encodedListId := t.String()
	return SendInternalEmail(r, email, "b55f71f4-8f0a-4540-a2b5-d74ee5249da1", "Your list has been uploaded!", "{LIST_ID}", encodedListId)
}
