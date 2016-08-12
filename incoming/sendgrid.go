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

func SendGridHandler(w http.ResponseWriter, r *http.Request) {

}
