package incoming

import (
	"net/http"
)

type Event struct {
	SgMessageID string `json:"sg_message_id"`
	Email       string `json:"email"`
	Timestamp   int    `json:"timestamp"`
	SMTPID      string `json:"smtp-id,omitempty"`
	Event       string `json:"event"`
	Category    string `json:"category,omitempty"`
	URL         string `json:"url,omitempty"`
	AsmGroupID  int    `json:"asm_group_id,omitempty"`
}

type Bounce struct {
	Status       string   `json:"status"`
	SgEventID    string   `json:"sg_event_id"`
	SgMessageID  string   `json:"sg_message_id"`
	Event        string   `json:"event"`
	Email        string   `json:"email"`
	Timestamp    int      `json:"timestamp"`
	SMTPID       string   `json:"smtp-id"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	AsmGroupID int    `json:"asm_group_id"`
	Reason     string `json:"reason"`
	Type       string `json:"type"`
	IP         string `json:"ip"`
	TLS        string `json:"tls"`
	CertErr    string `json:"cert_err"`
}

func SendGridHandler(w http.ResponseWriter, r *http.Request) {

}
