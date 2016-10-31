package emails

import (
	"net/http"
	"net/url"
	"strings"

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

	return emails.SendInternalEmail(r, email, "a64e454c-19d5-4bba-9cef-bd185e7c9b0b", "Thanks for signing up!", emailSubstitutes, 0)
}

func SendWelcomeEmail(r *http.Request, email models.Email) (bool, string, error) {
	emailSubstitutes := []emails.EmailSubstitute{}
	return emails.SendInternalEmail(r, email, "89d03e1c-6f9a-4820-8beb-db8b9856e7de", "Welcome To NewsAI Tabulae!", emailSubstitutes, 0)
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

	return emails.SendInternalEmail(r, email, "434520df-7773-424a-8e4a-8a6bf1e24441", "Reset your NewsAI password!", emailSubstitutes, 0)
}

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

func SendListUploadedEmail(r *http.Request, email models.Email, listId string) (bool, string, error) {
	// Adding the confirmation code for emails
	t := &url.URL{Path: listId}
	encodedListId := t.String()

	emailSubstitutes := []emails.EmailSubstitute{}
	emailSubstitute := emails.EmailSubstitute{}
	emailSubstitute.Name = "{LIST_ID}"
	emailSubstitute.Code = encodedListId
	emailSubstitutes = append(emailSubstitutes, emailSubstitute)

	return emails.SendInternalEmail(r, email, "b55f71f4-8f0a-4540-a2b5-d74ee5249da1", "Your list has been uploaded!", emailSubstitutes, 0)
}

func SendInvoiceEmail(r *http.Request, email models.Email, plan string, duration string, billDate string, billAmount string, paidAmount string) (bool, string, error) {
	emailSubstitutes := []emails.EmailSubstitute{}

	planSubstitute := emails.EmailSubstitute{}
	planSubstitute.Name = "{PLAN}"
	planSubstitute.Code = plan
	emailSubstitutes = append(emailSubstitutes, planSubstitute)

	durationSubstitute := emails.EmailSubstitute{} // a monthly or an annual
	durationSubstitute.Name = "{DURATION}"
	durationSubstitute.Code = duration
	emailSubstitutes = append(emailSubstitutes, durationSubstitute)

	billDateSubstitute := emails.EmailSubstitute{}
	billDateSubstitute.Name = "{BILLDATE}"
	billDateSubstitute.Code = billDate
	emailSubstitutes = append(emailSubstitutes, billDateSubstitute)

	billAmountSubstitute := emails.EmailSubstitute{}
	billAmountSubstitute.Name = "{BILLAMOUNT}"
	billAmountSubstitute.Code = billAmount
	emailSubstitutes = append(emailSubstitutes, billAmountSubstitute)

	planAmountSubstitute := emails.EmailSubstitute{}
	planAmountSubstitute.Name = "{PAIDAMOUNT}"
	planAmountSubstitute.Code = paidAmount
	emailSubstitutes = append(emailSubstitutes, planAmountSubstitute)

	return emails.SendInternalEmail(r, email, "743f5023-b9df-4d08-be22-611e56c01191", "Welcome to NewsAI Premium!", emailSubstitutes, 0)
}
