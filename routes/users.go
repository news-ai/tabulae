package routes

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/notifications"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

func handleUserActions(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "token":
			return notifications.GetUserToken(c, r)
		case "confirm-email":
			return api.BaseSingleResponseHandler(controllers.ConfirmAddEmailToUser(c, r, id))
		case "plan-details":
			return api.BaseSingleResponseHandler(controllers.GetUserPlanDetails(c, r, id))
		}
	case "POST":
		switch action {
		case "feedback":
			return api.BaseSingleResponseHandler(controllers.FeedbackFromUser(c, r, id))
		case "add-email":
			return api.BaseSingleResponseHandler(controllers.AddEmailToUser(c, r, id))
		case "remove-email":
			return api.BaseSingleResponseHandler(controllers.RemoveEmailFromUser(c, r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleUser(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetUser(c, r, id))
	case "PATCH":
		return api.BaseSingleResponseHandler(controllers.UpdateUser(c, r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleUsers(c context.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, err := controllers.GetUsers(c, r)
		return api.BaseResponseHandler(val, included, count, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the users.
func UsersHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleUsers(c, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func UserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleUser(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
	}
	return
}

func UserActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	action := ps.ByName("action")

	val, err := handleUserActions(c, r, id, action)

	if action == "confirm-email" {
		http.Redirect(w, r, "https://tabulae.newsai.co/settings", 302)
		return
	}

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "User handling error", err.Error())
	}
	return
}
