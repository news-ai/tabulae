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

func getUserBilling(c context.Context, r *http.Request, user models.User) (models.Billing, error) {
	ks, err := datastore.NewQuery("Billing").Filter("CreatedBy =", user.Id).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Billing{}, err
	}

	var billing []models.Billing
	billing = make([]models.Billing, len(ks))
	err = nds.GetMulti(c, ks, billing)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Billing{}, err
	}

	if len(billing) > 0 {
		billing[0].Format(ks[0], "billing")
		return billing[0], nil
	}
	return models.Billing{}, errors.New("No billing for this user")
}
