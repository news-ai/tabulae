package billing

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"

	"github.com/news-ai/tabulae/models"
)

type Card struct {
	LastFour  string
	IsDefault bool
	Brand     stripe.CardBrand
}

type StripeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func GetUserCards(r *http.Request, user models.User, userBilling *models.Billing) ([]Card, error) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	customer, err := sc.Customers.Get(userBilling.StripeId, nil)
	if err != nil {
		var stripeError StripeError
		err = json.Unmarshal([]byte(err.Error()), &stripeError)
		if err != nil {
			return []Card{}, errors.New("We had an error getting your user")
		}

		log.Errorf(c, "%v", err)
		return []Card{}, errors.New(stripeError.Message)
	}

	cards := []Card{}
	for i := 0; i < len(customer.Sources.Values); i++ {
		newCard := Card{}
		newCard.IsDefault = customer.Sources.Values[i].Card.Default
		newCard.LastFour = customer.Sources.Values[i].Card.LastFour
		newCard.Brand = customer.Sources.Values[i].Card.Brand
		cards = append(cards, newCard)
	}

	return cards, nil
}

func AddPaymentsToCustomer(r *http.Request, user models.User, userBilling *models.Billing, stripeToken string) error {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	params := &stripe.CustomerParams{}
	params.SetSource(stripeToken)

	_, err := sc.Customers.Update(
		userBilling.StripeId,
		params,
	)

	if err != nil {
		var stripeError StripeError
		err = json.Unmarshal([]byte(err.Error()), &stripeError)
		if err != nil {
			return errors.New("We had an error getting your user")
		}

		log.Errorf(c, "%v", err)
		return errors.New(stripeError.Message)
	}

	newCustomer, err := sc.Customers.Get(userBilling.StripeId, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return err
	}

	userBilling.CardsOnFile = []string{}
	for i := 0; i < len(newCustomer.Sources.Values); i++ {
		userBilling.CardsOnFile = append(userBilling.CardsOnFile, newCustomer.Sources.Values[i].ID)
	}
	userBilling.Save(c)

	return nil
}
