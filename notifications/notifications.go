package notifications

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/channel"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type Notification struct {
	Message string `json:"message"`
}

func SendNotification(r *http.Request, notification Notification) error {
	c := appengine.NewContext(r)

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	for i := 0; i < len(currentUser.TokenIds); i++ {
		err = channel.SendJSON(c, currentUser.TokenIds[i], notification)

		// Log the error but continue sending the notification to other clients
		// Future: remove the connection from the array if it has multiple errors
		if err != nil {
			log.Errorf(c, "%v", err)
		}
	}

	return nil
}

func UserConnect(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	token := r.FormValue("from")
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	currentUser.TokenIds = append(currentUser.TokenIds, token)
	currentUser.Save(c)

	w.WriteHeader(200)
	return
}

func UserDisconnect(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	token := r.FormValue("from")
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	for i := 0; i < len(currentUser.TokenIds); i++ {
		if currentUser.TokenIds[i] == token {
			currentUser.TokenIds = append(currentUser.TokenIds[:i], currentUser.TokenIds[i+1:]...)
		}
	}

	currentUser.Save(c)

	w.WriteHeader(200)
	return
}

func GetUserToken(c context.Context, r *http.Request) (interface{}, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	token, err := channel.Create(c, strconv.FormatInt(currentUser.Id, 10))
	if err != nil {
		log.Errorf(c, "channel.Create: %v", err)
		return nil, err
	}

	tokenResponse := TokenResponse{}
	tokenResponse.Token = token
	return tokenResponse, nil
}
