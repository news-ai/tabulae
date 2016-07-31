package emails

import (
	"net/http"
	"os"

	"github.com/news-ai/tabulae/models"

	"appengine"
	"appengine/urlfetch"

	"gopkg.in/sendgrid/sendgrid-go.v2"
)

var sg = sendgrid.NewSendGridClient("newsaiorg", os.Getenv("SENDGRID_KEY"))

// Basically means we'll send an email through our platform
func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) {
	c := appengine.NewContext(r)
	sg.Client = urlfetch.Client(c)

	m := sendgrid.NewMail()
	m.AddTo(email.To)
	m.SetSubject("Thanks for signing up!")
	m.SetHTML(" ")
	m.SetText(" ")
	m.SetFrom("Abhi from NewsAI <abhi@newsai.org>")
	m.AddFilter("templates", "enable", "1")
	m.AddFilter("templates", "template_id", "a64e454c-19d5-4bba-9cef-bd185e7c9b0b")
	m.AddSubstitution("{CONFIRMATION_CODE}", confirmationCode)

	if err := sg.Send(m); err != nil {
		c.Infof("Couldn't send email: %v", err)
	}
}
