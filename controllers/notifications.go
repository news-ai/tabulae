package controllers

import (
	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/models"
)

/*
* Public methods
 */

/*
* Get methods
 */

/*
* Create methods
 */

func CreateNotificationForUser(c context.Context, currentUser models.User) (models.Notification, error) {
	notification := models.Notification{}
	_, err := notification.Create(c, currentUser)
	if err != nil {
		return notification, err
	}
	return notification, nil
}

func CreateNotificationObjectForUser(c context.Context, currentUser models.User, resourceName string, resourceId int64) (models.NotificationObject, error) {
	notificationObject := models.NotificationObject{}
	notificationObject.NoticationId = 0
	notificationObject.Object = resourceName
	notificationObject.ObjectId = resourceId

	_, err := notificationObject.Create(c, currentUser)
	if err != nil {
		return notificationObject, err
	}
	return notificationObject, nil
}
