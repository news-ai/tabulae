package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/utilities"
)

func generateTokenAndEmail(c context.Context, r *http.Request, email string) (models.UserInviteCode, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return *contact, nil, err
	}

	referralCode := models.UserInviteCode{}
	referralCode.Email = email
	referralCode.InviteCode = utilities.RandToken()
	referralCode.ReferralUser = currentUser.Id
	_, err = referralCode.Create(c, r)
	if err != nil {
		return models.UserInviteCode{}, err
	}

	// Email this person with the referral code
	// email

	return referralCode, nil
}
