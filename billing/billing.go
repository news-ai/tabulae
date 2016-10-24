package billing

import (
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

func CreateCustomer(r *http.Request, user models.User) error {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	params := &stripe.CustomerParams{
		Balance: 0,
		Email:   user.Email,
	}

	customer, err := sc.Customers.New(params)
	if err != nil {
		return err
	}

	user.SetStripeId(c, r, user, customer.ID, "", false, false)
	return nil
}

func GetUserCards(r *http.Request, user models.User, userBilling *models.Billing) ([]Card, error) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	customer, err := sc.Customers.Get(userBilling.StripeId, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Card{}, err
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
		log.Errorf(c, "%v", err)
		return err
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
