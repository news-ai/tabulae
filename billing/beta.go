package billing

import (
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"

	"github.com/news-ai/tabulae/models"
)

func CreateBetaCustomer(r *http.Request, user models.User) error {
	if user.StripeId == "" {
		c := appengine.NewContext(r)
		httpClient := urlfetch.Client(c)
		sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

		params := &stripe.CustomerParams{
			Balance:  0,
			Email:    user.Email,
			TrialEnd: 1477958399,
			Plan:     "silver",
		}

		customer, err := sc.Customers.New(params)
		if err != nil {
			return err
		}

		user.SetStripeId(c, customer.ID, "silver", true, true)
	}
	return nil
}