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
func SendConfirmationEmailNew(r *http.Request, email models.Email, user models.User) (bool, string, error) {
	c := appengine.NewContext(r)

	sendgrid.DefaultClient.HTTPClient = urlfetch.Client(c)

	userFullName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	emailFullName := strings.Join([]string{email.FirstName, email.LastName}, " ")

	from := mail.NewEmail(userFullName, user.Email)
	to := mail.NewEmail(emailFullName, email.To)
	content := mail.NewContent("text/html", email.Body)
	m := mail.NewV3MailInit(from, email.Subject, to, content)

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
