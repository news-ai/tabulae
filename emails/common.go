package emails

import (
	"net/mail"
	"os"

	"gopkg.in/sendgrid/sendgrid-go.v2"
)

var sg = sendgrid.NewSendGridClient(os.Getenv("SENDGRID_USER"), os.Getenv("SENDGRID_KEY"))
var fromNewsAIEmail = mail.Address{Name: "Abhi Agarwal", Address: "abhi@newsai.org"}
