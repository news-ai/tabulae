package notifications

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/channel"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/web/utilities"
)

type TokenResponse struct {
	ChannelToken string `json:"token"`
}

type Notification struct {
	Message string `json:"message"`
}

func SendNotification(r *http.Request, notification Notification, userId int64) error {
	c := appengine.NewContext(r)
	userTokens, err := controllers.GetTokensForUser(c, r, userId, true)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	for i := 0; i < len(userTokens); i++ {
		err = channel.SendJSON(c, userTokens[i].ChannelToken, notification)

		// Log the error but continue sending the notification to other clients
		// Future: remove the connection from the array if it has multiple errors
		if err != nil {
			log.Errorf(c, "%v", err)
		}
	}

	return err
}

func UserConnect(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// When user connects send them notifications from the past
	token := r.FormValue("from")
	validToken, err := controllers.GetToken(c, r, token)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	validToken.IsUsed = true
	validToken.Save(c)

	w.WriteHeader(200)
	return
}

func UserDisconnect(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	token := r.FormValue("from")
	validToken, err := controllers.GetToken(c, r, token)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}
	validToken.IsUsed = false
	validToken.Save(c)
	w.WriteHeader(200)
	return
}

func GetUserToken(c context.Context, r *http.Request) (interface{}, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	userTokens, err := controllers.GetTokensForUser(c, r, currentUser.Id, false)
	if err != nil {
		log.Errorf(c, "channel.Create: %v", err)
		return nil, err
	}

	tokenResponse := TokenResponse{}

	if len(userTokens) > 0 {
		singleToken := userTokens[0]
		tokenResponse.ChannelToken = singleToken.ChannelToken
	} else {
		randomString := strconv.FormatInt(currentUser.Id, 10)
		randomString = randomString + utilities.RandToken()

		channelToken, err := channel.Create(c, randomString)
		if err != nil {
			log.Errorf(c, "channel.Create: %v", err)
			return nil, err
		}

		userToken := models.UserToken{}
		userToken.CreatedBy = currentUser.Id
		userToken.Token = randomString
		userToken.ChannelToken = channelToken
		userToken.Create(c, r)

		tokenResponse.ChannelToken = channelToken
	}

	return tokenResponse, nil
}
