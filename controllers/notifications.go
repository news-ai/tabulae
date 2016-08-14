package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/news-ai/tabulae/models"
)

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single media list
func GetUserNotification(c context.Context, r *http.Request) (models.Notification, error) {
	notifications := []models.Notification{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		return models.Notification{}, err
	}

	ks, err := datastore.NewQuery("Notification").Filter("CreatedBy =", user.Id).GetAll(c, &notifications)
	if err != nil {
		return models.Notification{}, err
	}

	for i := 0; i < len(notifications); i++ {
		notifications[i].Id = ks[i].IntID()
		return notifications[0], nil
	}

	return models.Notification{}, errors.New("No notification for this user")
}

/*
* Create methods
 */

func CreateNotificationForUser(c context.Context, r *http.Request) (models.Notification, error) {
	notification := models.Notification{}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return models.Notification{}, err
	}

	_, err = notification.Create(c, currentUser)
	if err != nil {
		return notification, err
	}
	return notification, nil
}

func CreateNotificationObjectForUser(c context.Context, r *http.Request, resourceName string, resourceId int64) (models.NotificationObject, error) {
	notificationObject := models.NotificationObject{}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return models.NotificationObject{}, err
	}

	userNotification, err := GetUserNotification(c, r)
	if err != nil {
		return models.NotificationObject{}, err
	}

	notificationObject.NoticationId = userNotification.Id
	notificationObject.Object = resourceName
	notificationObject.ObjectId = resourceId

	_, err = notificationObject.Create(c, currentUser)
	if err != nil {
		return notificationObject, err
	}
	return notificationObject, nil
}
