package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
)

/*
* Private methods
 */

/*
* Get methods
 */

// Get a user notification
func getUserNotification(c context.Context, r *http.Request) (models.Notification, error) {
	notifications := []models.Notification{}
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Notification{}, err
	}

	ks, err := datastore.NewQuery("Notification").Filter("CreatedBy =", user.Id).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Notification{}, err
	}

	notifications = make([]models.Notification, len(ks))
	err = nds.GetMulti(c, ks, notifications)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Notification{}, err
	}

	for i := 0; i < len(notifications); i++ {
		notifications[i].Id = ks[i].IntID()
		return notifications[0], nil
	}

	return models.Notification{}, errors.New("No notification for this user")
}

func getUserNotificationObjects(c context.Context, r *http.Request) ([]models.NotificationObject, error) {
	notificationObjects := []models.NotificationObject{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.NotificationObject{}, err
	}

	ks, err := datastore.NewQuery("NotificationObject").Filter("CreatedBy =", user.Id).Limit(1).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.NotificationObject{}, err
	}

	notificationObjects = make([]models.NotificationObject, len(ks))
	err = nds.GetMulti(c, ks, notificationObjects)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.NotificationObject{}, err
	}

	for i := 0; i < len(notificationObjects); i++ {
		notificationObjects[i].Id = ks[i].IntID()
	}
	return notificationObjects, nil
}

/*
* Create methods
 */

func createNotificationChange(c context.Context, r *http.Request, notificationObjectId int64, verb, actor string) (models.NotificationChange, error) {
	notificationChange := models.NotificationChange{}
	notificationChange.NoticationObjectId = notificationObjectId
	notificationChange.Verb = verb
	notificationChange.Actor = actor

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.NotificationChange{}, err
	}

	notificationChange.Create(c, r, user)

	return notificationChange, nil
}

/*
* Filter methods
 */

func filterNotificationObject(c context.Context, r *http.Request, resourceName string, resourceId int64) (models.NotificationObject, error) {
	// Get notification by resource name
	notificationObjects := []models.NotificationObject{}

	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.NotificationObject{}, err
	}

	ks, err := datastore.NewQuery("NotificationObject").Filter("CreatedBy =", user.Id).Filter("Object =", resourceName).Filter("ObjectId =", resourceId).GetAll(c, &notificationObjects)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.NotificationObject{}, err
	}
	if len(notificationObjects) > 0 {
		notificationObjects[0].Id = ks[0].IntID()
		return notificationObjects[0], nil
	}
	return models.NotificationObject{}, errors.New("No notification object by this Object")
}

/*
* Public methods
 */

/*
* Create methods
 */

func CreateNotificationForUser(c context.Context, r *http.Request) (models.Notification, error) {
	notification := models.Notification{}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Notification{}, err
	}

	_, err = notification.Create(c, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return notification, err
	}
	return notification, nil
}

func CreateNotificationObjectForUser(c context.Context, r *http.Request, resourceName string, resourceId int64) (models.NotificationObject, error) {
	notificationObject := models.NotificationObject{}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.NotificationObject{}, err
	}

	userNotification, err := getUserNotification(c, r)
	if err != nil {
		userNotification, err = CreateNotificationForUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.NotificationObject{}, err
		}
	}

	notificationObject.NoticationId = userNotification.Id
	notificationObject.Object = resourceName
	notificationObject.ObjectId = resourceId

	_, err = notificationObject.Create(c, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return notificationObject, err
	}
	return notificationObject, nil
}

/*
* Filter methods
 */

func FilterNotificationObjectByObject(c context.Context, r *http.Request, resourceName string, resourceId int64) (models.NotificationObject, error) {
	// Get the id of a notification object for a user
	notifiation, err := filterNotificationObject(c, r, resourceName, resourceId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.NotificationObject{}, err
	}
	return notifiation, nil
}

/*
* Action methods
 */

func LogNotificationForResource(c context.Context, r *http.Request, resourceName string, resourceId int64, verb, actor string) (models.NotificationChange, error) {
	notificationObject, err := FilterNotificationObjectByObject(c, r, resourceName, resourceId)
	if err != nil {
		log.Errorf(c, "%v", err)
		notificationObject, err = CreateNotificationObjectForUser(c, r, resourceName, resourceId)
	}
	return createNotificationChange(c, r, notificationObject.Id, verb, actor)
}
