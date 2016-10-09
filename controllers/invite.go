package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/mail"

	"golang.org/x/net/context"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/utilities"
)

func generateTokenAndEmail(c context.Context, r *http.Request, email string) (models.UserInviteCode, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	validEmail, err := mail.ParseAddress(email)
	if err != nil {
		invalidEmailError := errors.New("Email user has entered is incorrect")
		log.Errorf(c, "%v", invalidEmailError)
		return models.UserInviteCode{}, invalidEmailError
	}

	// Get the Contact by id
	ks, err := datastore.NewQuery("UserInviteCode").Filter("Email =", validEmail.Address).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	if len(ks) > 0 {
		invitedAlreadyError := errors.New("User has already been invited to the NewsAI platform")
		log.Errorf(c, "%v", invitedAlreadyError)
		return models.UserInviteCode{}, invitedAlreadyError
	}

	// Check if the user is already a part of the platform
	_, err = GetUserByEmail(c, validEmail.Address)
	if err == nil {
		userExistsError := errors.New("User already exists on the NewsAI platform")
		log.Errorf(c, "%v", userExistsError)
		return models.UserInviteCode{}, userExistsError
	}

	referralCode := models.UserInviteCode{}
	referralCode.Email = email
	referralCode.InviteCode = utilities.RandToken()
	_, err = referralCode.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	// Email this person with the referral code
	emailInvitaiton, _ := CreateEmailInternal(r, validEmail.Address, "", "")
	emails.SendInvitationEmail(r, emailInvitaiton, referralCode.InviteCode)

	return referralCode, nil
}

func GetInviteFromInvitationCode(c context.Context, r *http.Request, invitationCode string) (models.UserInviteCode, error) {
	currentId, err := utilities.StringIdToInt(invitationCode)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	// Get the agency by id
	var userInviteCode models.UserInviteCode
	userInviteCodeId := datastore.NewKey(c, "UserInviteCode", "", currentId, nil)
	err = nds.Get(c, userInviteCodeId, &userInviteCode)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	return userInviteCode, nil
}

func CreateInvite(c context.Context, r *http.Request) (models.UserInviteCode, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var invite models.Invite
	err := decoder.Decode(buf, &invite)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, nil, err
	}

	userInvite, err := generateTokenAndEmail(c, r, invite.Email)
	if err != nil {
		return models.UserInviteCode{}, nil, err
	}

	userInvite.Type = "invites"

	return userInvite, nil, nil
}
