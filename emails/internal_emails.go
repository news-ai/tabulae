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
	return emails.SendInternalEmail(r, email, "a64e454c-19d5-4bba-9cef-bd185e7c9b0b", "Thanks for signing up!", "{CONFIRMATION_CODE}", encodedConfirmationCode)
}

func SendResetEmail(r *http.Request, email models.Email, resetPasswordCode string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: resetPasswordCode}
	encodedResetCode := t.String()
	return emails.SendInternalEmail(r, email, "434520df-7773-424a-8e4a-8a6bf1e24441", "Reset your NewsAI password!", "{RESET_CODE}", encodedResetCode)
}

func SendListUploadedEmail(r *http.Request, email models.Email, listId string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: listId}
	encodedListId := t.String()
	return emails.SendInternalEmail(r, email, "b55f71f4-8f0a-4540-a2b5-d74ee5249da1", "Your list has been uploaded!", "{LIST_ID}", encodedListId)
}
