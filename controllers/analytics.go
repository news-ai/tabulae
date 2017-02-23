package controllers

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

/*
* Analytics methods
 */
func GetNumberOfEmailsCreatedMonth(c context.Context, r *http.Request) (int, error) {
	_, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	timeNow := time.Now()

	beginningOfMonthTime := time.Date(timeNow.Year(), timeNow.Month(), 1, 0, 0, 0, 0, timeNow.Location())

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("Created >=", beginningOfMonthTime)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	return len(ks), nil
}

func GetNumberOfEmailsCreatedToday(c context.Context, r *http.Request) (int, error) {
	_, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	timeNow := time.Now()

	morningTime := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("Created >=", morningTime).Filter("Cancel =", false).Filter("IsSent =", true)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	return len(ks), nil
}

func GetNumberOfScheduledEmails(c context.Context, r *http.Request) (int, error) {
	_, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	return len(ks), nil
}
