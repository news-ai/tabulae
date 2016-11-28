package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/encrypt"
	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
)

func getEmailSetting(c context.Context, r *http.Request, id int64) (models.EmailSetting, error) {
	if id == 0 {
		return models.EmailSetting{}, errors.New("datastore: no such entity")
	}
	// Get the email by id
	var emailSetting models.EmailSetting
	emailSettingId := datastore.NewKey(c, "EmailSetting", "", id, nil)
	err := nds.Get(c, emailSettingId, &emailSetting)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.EmailSetting{}, err
	}

	if !emailSetting.Created.IsZero() {
		emailSetting.Format(emailSettingId, "emailsettings")

		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.EmailSetting{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(emailSetting.CreatedBy, user.Id) {
			return models.EmailSetting{}, errors.New("Forbidden")
		}

		return emailSetting, nil
	}
	return models.EmailSetting{}, errors.New("No email setting by this id")
}

func GetEmailSetting(c context.Context, r *http.Request, id string) (models.EmailSetting, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.EmailSetting{}, nil, err
	}

	emailSetting, err := getEmailSetting(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.EmailSetting{}, nil, err
	}

	return emailSetting, nil, nil
}

func GetEmailSettings(c context.Context, r *http.Request) ([]models.EmailSetting, interface{}, int, error) {
	emailSettings := []models.EmailSetting{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.EmailSetting{}, nil, 0, err
	}

	query := datastore.NewQuery("EmailSetting").Filter("CreatedBy =", user.Id)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.EmailSetting{}, nil, 0, err
	}

	emailSettings = make([]models.EmailSetting, len(ks))
	err = nds.GetMulti(c, ks, emailSettings)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.EmailSetting{}, nil, 0, err
	}

	for i := 0; i < len(emailSettings); i++ {
		emailSettings[i].Format(ks[i], "emailsettings")
	}

	return emailSettings, nil, len(emailSettings), nil
}

/*
* Create methods
 */

func AddUserEmail(c context.Context, r *http.Request) (models.User, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	var userEmailSettings models.UserEmailSetting
	err = decoder.Decode(buf, &userEmailSettings)
	if err != nil {
		return models.User{}, nil, err
	}

	userPw, err := encrypt.EncryptString(userEmailSettings.SMTPPassword, os.Getenv("SECRETKEYEMAILPW"))
	if err != nil {
		return models.User{}, nil, err
	}

	currentUser.SMTPPassword = []byte(userPw)
	SaveUser(c, r, &currentUser)

	return currentUser, nil, nil
}

func CreateEmailSettings(c context.Context, r *http.Request) (models.EmailSetting, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.EmailSetting{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	var emailSettings models.EmailSetting
	err = decoder.Decode(buf, &emailSettings)
	if err != nil {
		return models.EmailSetting{}, nil, err
	}

	// Create email setting
	_, err = emailSettings.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.EmailSetting{}, nil, err
	}

	emailSettings.Type = "emailsettings"

	currentUser.EmailSetting = emailSettings.Id
	SaveUser(c, r, &currentUser)
	return emailSettings, nil, nil
}

func VerifyEmailSetting(c context.Context, r *http.Request, id string) (models.EmailSetting, interface{}, error) {
	return models.EmailSetting{}, nil, nil
}