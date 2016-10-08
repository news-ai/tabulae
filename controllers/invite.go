package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"github.com/pquerna/ffjson/ffjson"

	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/utilities"
)

func generateTokenAndEmail(c context.Context, r *http.Request, email string) (models.UserInviteCode, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	// Check if the user is already a part of the platform
	_, err = GetUserByEmail(c, email)
	if err == nil {
		userExistsError := errors.New("User already exists on the NewsAI platform")
		log.Errorf(c, "%v", userExistsError)
		return models.UserInviteCode{}, userExistsError
	}

	referralCode := models.UserInviteCode{}
	referralCode.Email = email
	referralCode.InviteCode = utilities.RandToken()
	referralCode.ReferralUser = currentUser.Id
	_, err = referralCode.Create(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	// Email this person with the referral code
	// email

	return referralCode, nil
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

	return userInvite, nil, nil
}
