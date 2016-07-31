package emails

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/news-ai/tabulae/models"

	"appengine"
	"appengine/mail"
)

// Basically means we'll send an email through our platform
func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) {
	c := appengine.NewContext(r)

	userEmail, err := models.GetEmailUser(c, strconv.FormatInt(email.To[0], 10))
	if err != nil {
		c.Infof("Couldn't send email: %v", err)
		return
	}

	msg := &mail.Message{
		Sender:   "Abhi from NewsAI <abhi@newsai.org>",
		To:       []string{userEmail.To},
		Subject:  "Thanks for signing up!",
		HTMLBody: strings.Replace(confirmMessage, "{CONFIRMATION_CODE}", confirmationCode, -1),
	}
	if err := mail.Send(c, msg); err != nil {
		c.Infof("Couldn't send email: %v", err)
	}
}
