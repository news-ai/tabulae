package emails

import (
	"net/http"
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
func SendEmail(r *http.Request, email models.Email, user models.User) (bool, string) {
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
		log.Errorf(c, "%v", response)
		return false, ""
	}

	log.Infof(c, "%v", response)
	emailId := request.Headers["X-Message-Id"]
	return true, emailId
}
