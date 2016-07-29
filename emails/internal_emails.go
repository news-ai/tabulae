package emails

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

// Basically means we'll send an email through our platform
func SendConfirmationEmail(r *http.Request, email Email, confirmationCode string) {
	c := appengine.NewContext(r)
	msg := &mail.Message{
		Sender:  "Abhi from NewsAI <abhi@newsai.org>",
		To:      email.To,
		Subject: "Thanks for signing up!",
		Body:    fmt.Sprintf(confirmMessage, confirmationCode),
	}
	if err := mail.Send(c, msg); err != nil {
		log.Errorf(c, "Couldn't send email: %v", err)
	}
}
