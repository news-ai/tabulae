package emails

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/emails"
)

func SendInvitationEmail(r *http.Request, email models.Email, currentUser models.User, invitationCode string, personalMessage string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: invitationCode}
	encodedInvitationCode := t.String()

	emailSubstitutes := []emails.EmailSubstitute{}

	// Invitation code
	emailSubstituteInvitationCode := emails.EmailSubstitute{}
	emailSubstituteInvitationCode.Name = "{INVITATION_CODE}"
	emailSubstituteInvitationCode.Code = encodedInvitationCode
	emailSubstitutes = append(emailSubstitutes, emailSubstituteInvitationCode)

	emailSubstituteNewUserEmail := emails.EmailSubstitute{}
	emailSubstituteNewUserEmail.Name = "{NEWUSER_EMAIL}"
	emailSubstituteNewUserEmail.Code = email.To
	emailSubstitutes = append(emailSubstitutes, emailSubstituteNewUserEmail)

	// Personal Message
	if personalMessage != "" {
		emailSubstitutePersonalMessage := emails.EmailSubstitute{}
		emailSubstitutePersonalMessage.Name = "{PERSONAL_MESSAGE}"
		emailSubstitutePersonalMessage.Code = personalMessage
		emailSubstitutes = append(emailSubstitutes, emailSubstitutePersonalMessage)
	}

	emailSubstituteFullName := emails.EmailSubstitute{}
	emailSubstituteFullName.Name = "{CURRENTUSER_FULL_NAME}"
	emailSubstituteFullName.Code = strings.Join([]string{currentUser.FirstName, currentUser.LastName}, " ")
	emailSubstitutes = append(emailSubstitutes, emailSubstituteFullName)

	emailSubstituteEmail := emails.EmailSubstitute{}
	emailSubstituteEmail.Name = "{CURRENTUSER_EMAIL}"
	emailSubstituteEmail.Code = currentUser.Email
	emailSubstitutes = append(emailSubstitutes, emailSubstituteEmail)

	return emails.SendInternalEmail(r, email, "47644933-3501-4f4c-a710-1c993c9925b8", "You've been invited to NewsAI Tabulae!", emailSubstitutes, 0)
}
