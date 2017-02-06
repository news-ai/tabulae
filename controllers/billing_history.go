package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getBillingHistory(c context.Context, id int64) (models.BillingHistory, error) {
	if id == 0 {
		return models.BillingHistory{}, errors.New("datastore: no such entity")
	}
	// Get the publication details by id
	var billingHistory models.BillingHistory
	billingHistoryId := datastore.NewKey(c, "BillingHistory", "", id, nil)

	err := nds.Get(c, billingHistoryId, &billingHistory)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.BillingHistory{}, err
	}

	if !billingHistory.Created.IsZero() {
		billingHistory.Format(billingHistoryId, "billinghistories")
		return billingHistory, nil
	}
	return models.BillingHistory{}, errors.New("No billing histoy by this id")
}
