package routes

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

func handleEmailSettingAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "verify":
			return api.BaseSingleResponseHandler(controllers.VerifyEmailSetting(c, r, id))
		case "details":
			return api.BaseSingleResponseHandler(controllers.GetEmailSettingDetails(c, r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleEmailSetting(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetEmailSetting(c, r, id))
	case "POST":
		switch id {
		case "add-email":
			return api.BaseSingleResponseHandler(controllers.AddUserEmail(c, r))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleEmailSettings(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, err := controllers.GetEmailSettings(c, r)
		return api.BaseResponseHandler(val, included, count, err, r)
	case "POST":
		return api.BaseSingleResponseHandler(controllers.CreateEmailSettings(c, r))
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func EmailSettingsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleEmailSettings(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Email setting handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /emailsettings/<id> route.
func EmailSettingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleEmailSetting(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Email setting handling error", err.Error())
	}
	return
}

func EmailSettingActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	action := ps.ByName("action")
	val, err := handleEmailSettingAction(c, r, id, action)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Email setting handling error", err.Error())
	}
	return
}
