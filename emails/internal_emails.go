package emails

import (
	"net/http"
	"strings"

	"github.com/news-ai/tabulae/models"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

// Basically means we'll send an email through our platform
func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) {
	c := appengine.NewContext(r)
	msg := &mail.Message{
		Sender:   "Abhi from NewsAI <abhi@newsai.org>",
		To:       []string{email.To[0].To},
		Subject:  "Thanks for signing up!",
		HTMLBody: strings.Replace(confirmMessage, "{CONFIRMATION_CODE}", confirmationCode, -1),
	}
	if err := mail.Send(c, msg); err != nil {
		log.Errorf(c, "Couldn't send email: %v", err)
	}
}
