package emails

import (
	"net/http"
	"net/url"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/emails"
)

func SendConfirmationEmail(r *http.Request, email models.Email, confirmationCode string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: confirmationCode}
	encodedConfirmationCode := t.String()

	emailSubstitutes := []emails.EmailSubstitute{}
	emailSubstitute := emails.EmailSubstitute{}
	emailSubstitute.Name = "{CONFIRMATION_CODE}"
	emailSubstitute.Code = encodedConfirmationCode
	emailSubstitutes = append(emailSubstitutes, emailSubstitute)

	return emails.SendInternalEmail(r, email, "a64e454c-19d5-4bba-9cef-bd185e7c9b0b", "Thanks for signing up!", emailSubstitutes)
}

func SendWelcomeEmail(r *http.Request, email models.Email) (bool, string, error) {
	emailSubstitutes := []emails.EmailSubstitute{}
	return emails.SendInternalEmail(r, email, "89d03e1c-6f9a-4820-8beb-db8b9856e7de", "Welcome To NewsAI Tabulae!", emailSubstitutes)
}

func SendResetEmail(r *http.Request, email models.Email, resetPasswordCode string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: resetPasswordCode}
	encodedResetCode := t.String()

	emailSubstitutes := []emails.EmailSubstitute{}
	emailSubstitute := emails.EmailSubstitute{}
	emailSubstitute.Name = "{RESET_CODE}"
	emailSubstitute.Code = encodedResetCode
	emailSubstitutes = append(emailSubstitutes, emailSubstitute)

	return emails.SendInternalEmail(r, email, "434520df-7773-424a-8e4a-8a6bf1e24441", "Reset your NewsAI password!", emailSubstitutes)
}

func SendInvitationEmail(r *http.Request, email models.Email, invitationCode string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: invitationCode}
	encodedInvitationCode := t.String()

	emailSubstitutes := []emails.EmailSubstitute{}
	emailSubstitute := emails.EmailSubstitute{}
	emailSubstitute.Name = "{INVITATION_CODE}"
	emailSubstitute.Code = encodedInvitationCode
	emailSubstitutes = append(emailSubstitutes, emailSubstitute)

	return emails.SendInternalEmail(r, email, "47644933-3501-4f4c-a710-1c993c9925b8", "You've been invited to NewsAI Tabulae!", emailSubstitutes)
}

func SendListUploadedEmail(r *http.Request, email models.Email, listId string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: listId}
	encodedListId := t.String()

	emailSubstitutes := []emails.EmailSubstitute{}
	emailSubstitute := emails.EmailSubstitute{}
	emailSubstitute.Name = "{LIST_ID}"
	emailSubstitute.Code = encodedListId
	emailSubstitutes = append(emailSubstitutes, emailSubstitute)

	return emails.SendInternalEmail(r, email, "b55f71f4-8f0a-4540-a2b5-d74ee5249da1", "Your list has been uploaded!", emailSubstitutes)
}
