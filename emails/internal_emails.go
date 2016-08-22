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
func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) (bool, string, error) {
	c := appengine.NewContext(r)

	sendgrid.DefaultClient.HTTPClient = urlfetch.Client(c)

	emailFullName := strings.Join([]string{email.FirstName, email.LastName}, " ")

	from := mail.NewEmail("Abhi from NewsAI", "abhi@newsai.org")
	to := mail.NewEmail(emailFullName, email.To)
	content := mail.NewContent("text/html", email.Body)
	m := mail.NewV3MailInit(from, "Thanks for signing up!", to, content)
	m.SetTemplateID("a64e454c-19d5-4bba-9cef-bd185e7c9b0b")

	// personalization := mail.Personalization{}
	// personalization.Substitutions

	// m.AddPersonalizations()
	// m.AddS

	//url.QueryEscape(confirmationCode)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)

	// Send the actual mail here
	response, err := sendgrid.API(request)
	if err != nil {
		log.Errorf(c, "error: %v", err)
		return false, "", err
	}

	// 	m.AddSubstitution("{CONFIRMATION_CODE}", url.QueryEscape(confirmationCode))

	emailId := ""
	if len(response.Headers["X-Message-Id"]) > 0 {
		emailId = response.Headers["X-Message-Id"][0]
	}
	return true, emailId, nil
}
