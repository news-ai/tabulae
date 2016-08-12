package incoming

import (
	"net/http"
)

type SendGridEvent struct {
	SgMessageID string `json:"sg_message_id"`
	Email       string `json:"email"`
	Timestamp   int    `json:"timestamp"`
	SMTPID      string `json:"smtp-id,omitempty"`
	Event       string `json:"event"`
	Category    string `json:"category,omitempty"`
	URL         string `json:"url,omitempty"`
	AsmGroupID  int    `json:"asm_group_id,omitempty"`
}

type OpenEvent struct {
	SendGridEvent

	IP           string   `json:"ip"`
	SgEventID    string   `json:"sg_event_id"`
	Useragent    string   `json:"useragent"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
}

type DeliveredEvent struct {
	SendGridEvent

	Response     string   `json:"response"`
	SgEventID    string   `json:"sg_event_id"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	IP      string `json:"ip"`
	TLS     string `json:"tls"`
	CertErr string `json:"cert_err"`
}

type ClickEvent struct {
	SendGridEvent

	SgEventID string `json:"sg_event_id"`
	IP        string `json:"ip"`
	Useragent string `json:"useragent"`
	URLOffset struct {
		Index int    `json:"index"`
		Type  string `json:"type"`
	} `json:"url_offset"`
	UniqueArgKey string   `json:"unique_arg_key"`
	Category     []string `json:"category"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
}

type BounceEvent struct {
	SendGridEvent

	Status       string `json:"status"`
	SgEventID    string `json:"sg_event_id"`
	UniqueArgKey string `json:"unique_arg_key"`
	Newsletter   struct {
		NewsletterUserListID string `json:"newsletter_user_list_id"`
		NewsletterID         string `json:"newsletter_id"`
		NewsletterSendID     string `json:"newsletter_send_id"`
	} `json:"newsletter"`
	Reason  string `json:"reason"`
	Type    string `json:"type"`
	IP      string `json:"ip"`
	TLS     string `json:"tls"`
	CertErr string `json:"cert_err"`
}

func SendGridHandler(w http.ResponseWriter, r *http.Request) {

}
