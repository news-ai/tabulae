package notifications

import (
	"net/http"
	"strconv"
	"time"

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
	ResourceId  int64  `json:"resourceid"`
	ResouceName string `json:"resourcename"`
	Verb        string `json:"verb"`

	Message string `json:"message"`
}

func (n *Notification) generateNotificationMessage(c context.Context, r *http.Request) {
	if n.ResouceName == "Email" {
		email, err := controllers.GetEmailById(c, r, n.ResourceId)
		if err != nil {
			log.Infof(c, "%v", err)
			n.Message = "One of your emails were opened!"
			return
		}
		if n.Verb == "OPENED" {
			n.Message = email.To + " opened your email!"
			return
		}
	}
}

func SendNotification(r *http.Request, notificationChanges []models.NotificationChange, userId int64) error {
	c := appengine.NewContext(r)
	// Set the current user logged in
	controllers.SetUser(c, r, userId)

	// Grab the user's tokens
	userTokens, err := controllers.GetTokensForUser(c, r, userId, true)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	notifications := []Notification{}
	for i := 0; i < len(notificationChanges); i++ {
		objectNotification, err := controllers.GetNotificationObjectById(c, r, notificationChanges[i].NoticationObjectId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return err
		}

		notification := Notification{}
		notification.ResouceName = objectNotification.Object
		notification.ResourceId = objectNotification.ObjectId
		notification.Verb = notificationChanges[i].Verb
		notification.generateNotificationMessage(c, r)
		notifications = append(notifications, notification)
	}

	for i := 0; i < len(userTokens); i++ {
		err = channel.SendJSON(c, userTokens[i].ChannelToken, notifications)

		// Log the error but continue sending the notification to other clients
		// Future: remove the connection from the array if it has multiple errors
		if err != nil {
			log.Errorf(c, "%v", err)
		}
	}

	// After sending mark everything as read

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

	notifications, err := controllers.GetUnreadNotificationsForUser(c, r, validToken.CreatedBy)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}
	if len(notifications) > 0 {
		err = SendNotification(r, notifications, validToken.CreatedBy)
		if err != nil {
			log.Errorf(c, "%v", err)
			w.WriteHeader(500)
			return
		}
	}

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

func generateChannelToken(c context.Context, currentUser models.User) (string, string, error) {
	randomString := strconv.FormatInt(currentUser.Id, 10)
	randomString = randomString + utilities.RandToken()

	channelToken, err := channel.Create(c, randomString)
	if err != nil {
		log.Errorf(c, "channel.Create: %v", err)
		return "", "", err
	}

	return randomString, channelToken, nil
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

		fiftenMinutes := singleToken.Updated.Add(time.Minute * 15)

		if time.Now().After(fiftenMinutes) {
			randomString, channelToken, err := generateChannelToken(c, currentUser)
			if err != nil {
				log.Errorf(c, "%v", err)
				return nil, err
			}
			singleToken.Token = randomString
			singleToken.ChannelToken = channelToken
			singleToken.Save(c)
		}

		tokenResponse.ChannelToken = singleToken.ChannelToken
	} else {
		randomString, channelToken, err := generateChannelToken(c, currentUser)
		if err != nil {
			log.Errorf(c, "%v", err)
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
