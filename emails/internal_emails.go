package emails

import (
	"net/http"
	"net/url"

	"github.com/news-ai/tabulae/models"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"gopkg.in/sendgrid/sendgrid-go.v2"
)

// Send an email confirmation to a new user
func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) {
	c := appengine.NewContext(r)
	sg.Client = urlfetch.Client(c)

	m := sendgrid.NewMail()
	m.AddTo(email.To)
	m.AddToName(email.FirstName + " " + email.LastName)
	m.SetSubject("Thanks for signing up!")
	m.SetHTML(" ")
	m.SetText(" ")
	m.SetFrom("Abhi from NewsAI <abhi@newsai.org>")
	m.SetReplyTo(fromNewsAIEmail.Name)
	m.SetReplyToEmail(&fromNewsAIEmail)
	m.AddFilter("templates", "enable", "1")
	m.AddFilter("templates", "template_id", "a64e454c-19d5-4bba-9cef-bd185e7c9b0b")
	m.AddSubstitution("{CONFIRMATION_CODE}", url.QueryEscape(confirmationCode))
	sg.Send(m)
}
