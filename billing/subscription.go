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

func AddPlanToUser(r *http.Request, user models.User, plan string, months int, coupon string, sp string) error {
	if user.StripeId != "" {
		c := appengine.NewContext(r)
		httpClient := urlfetch.Client(c)
		sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

		// Create new customer in Stripe
		params := &stripe.CustomerParams{
			Email:    user.Email,
			Plan:     plan,
			Quantity: 1 * months,
			Coupon:   coupon,
		}

		params.SetSource(sp)

		customer, err := sc.Customers.New(params)
		if err != nil {
			return err
		}

		user.SetStripeId(c, customer.ID, "", false, true)
	}
	return nil
}
