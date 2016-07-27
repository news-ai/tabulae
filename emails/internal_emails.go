package email

import (
	"bytes"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

// Basically means we'll send an email through our platform
func SendInternalEmail(r *http.Request, email Email) {
	c := appengine.NewContext(r)
	msg := &mail.Message{
		Sender:  "Abhi Agarwal <abhi@newsai.org>",
		To:      email.To,
		Subject: "Thanks for signing up!",
		Body:    fmt.Sprintf(confirmMessage),
	}
	if err := mail.Send(ctx, msg); err != nil {
		log.Errorf(ctx, "Couldn't send email: %v", err)
	}
}
