package emails

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/news-ai/tabulae/models"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"gopkg.in/sendgrid/sendgrid-go.v2"
)

// Send an email confirmation to a new user
func SendEmail(r *http.Request, email models.Email, user models.User) bool {
	c := appengine.NewContext(r)
	sg.Client = urlfetch.Client(c)

	userFullName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	var fromEmail = mail.Address{Name: userFullName, Address: user.Email}

	emailFullName := strings.Join([]string{email.FirstName, email.LastName}, " ")

	m := sendgrid.NewMail()
	m.AddTo(email.To)
	m.AddToName(emailFullName)
	m.SetSubject(email.Subject)
	m.SetHTML(email.Body)
	m.SetFrom(fromEmail.String())
	m.SetReplyTo(fromEmail.Name)
	m.SetReplyToEmail(&fromEmail)

	if err := sg.Send(m); err != nil {
		return false
	}
	return true
}
