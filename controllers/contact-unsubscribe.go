package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/models"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getUnsubscribedContact(c context.Context, id int64) (models.ContactUnsubscribe, error) {
	if id == 0 {
		return models.ContactUnsubscribe{}, errors.New("datastore: no such entity")
	}

	// Get the contactUnsubscribe by id
	var contactUnsubscribe models.ContactUnsubscribe
	contactUnsubscribeId := datastore.NewKey(c, "ContactUnsubscribe", "", id, nil)
	err := nds.Get(c, contactUnsubscribeId, &contactUnsubscribe)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.ContactUnsubscribe{}, err
	}

	if !contactUnsubscribe.Created.IsZero() {
		contactUnsubscribe.Format(contactUnsubscribeId, "unsubscribedcontacts")
		return contactUnsubscribe, nil
	}

	return models.ContactUnsubscribe{}, errors.New("No contact unsubscribed by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single agency
func GetUnsubscribedContacts(c context.Context, r *http.Request) ([]models.ContactUnsubscribe, interface{}, int, int, error) {
	// Now if user is not querying then check
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.ContactUnsubscribe{}, nil, 0, 0, err
	}

	if !user.IsAdmin {
		return []models.ContactUnsubscribe{}, nil, 0, 0, errors.New("Forbidden")
	}

	query := datastore.NewQuery("ContactUnsubscribe")
	query = controllers.ConstructQuery(query, r)

	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.ContactUnsubscribe{}, nil, 0, 0, err
	}

	var unsubscribedContacts []models.ContactUnsubscribe
	unsubscribedContacts = make([]models.ContactUnsubscribe, len(ks))
	err = nds.GetMulti(c, ks, unsubscribedContacts)
	if err != nil {
		log.Infof(c, "%v", err)
		return unsubscribedContacts, nil, 0, 0, err
	}

	for i := 0; i < len(unsubscribedContacts); i++ {
		unsubscribedContacts[i].Format(ks[i], "unsubscribedcontacts")
	}

	return unsubscribedContacts, nil, len(unsubscribedContacts), 0, nil
}
