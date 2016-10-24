package auth

import (
	"net/http"
	"net/url"
	"text/template"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/billing"
	"github.com/news-ai/tabulae/controllers"

	"github.com/gorilla/csrf"
)

func TrialPlanPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		if !user.IsActive {
			userBilling, err := controllers.GetUserBilling(c, r, user)

			// If the user has a user billing
			if err == nil {
				if userBilling.HasTrial && !userBilling.Expires.IsZero() {
					// If the user has already had a trial and has expired
					// This means: If userBilling Expire date is before the current time
					if userBilling.Expires.Before(time.Now()) {
						http.Redirect(w, r, "/api/billing", 302)
						return
					}

					// If the user has already had a trial but it has not expired
					if userBilling.Expires.After(time.Now()) {
						http.Redirect(w, r, "https://site.newsai.org/", 302)
						return
					}
				}
			}

			data := map[string]interface{}{
				"userEmail":      user.Email,
				csrf.TemplateTag: csrf.TemplateField(r),
			}

			t := template.New("trial.html")
			t, _ = t.ParseFiles("billing/trial.html")
			t.Execute(w, data)
		} else {
			// If the user is active then they don't need to start a free trial
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}
	}
}

func ChoosePlanPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		data := map[string]interface{}{
			"userEmail":      user.Email,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t := template.New("plans.html")
		t, _ = t.ParseFiles("billing/plans.html")
		t.Execute(w, data)
	}
}

func ChooseTrialPlanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		plan := r.FormValue("plan")

		// To check if there is a user logged in
		user, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		billingId, err := billing.AddFreeTrialToUser(r, user, plan)
		user.IsActive = true
		user.BillingId = billingId
		user.Save(c)

		// If there was an error creating this person's trial
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}

		// If not then their is now probably successful so we redirect them back
		returnURL := "https://site.newsai.org/"
		session, _ := Store.Get(r, "sess")
		if session.Values["next"] != nil {
			returnURL = session.Values["next"].(string)
		}
		u, err := url.Parse(returnURL)

		// If there's an error in parsing the return value
		// then returning it.
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, returnURL, 302)
			return
		}

		// This would be a bug since they should not be here if they
		// are a firstTimeUser. But we'll allow it to help make
		// experience normal.
		if user.LastLoggedIn.IsZero() {
			q := u.Query()
			q.Set("firstTimeUser", "true")
			u.RawQuery = q.Encode()
			user.ConfirmLoggedIn(c)
		}

		http.Redirect(w, r, u.String(), 302)
		return
	}
}

func BillingPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		userBilling, err := controllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			switch userBilling.StripePlanId {
			case "bronze":
				userBilling.StripePlanId = "Personal"
			case "silver":
				userBilling.StripePlanId = "Business"
			case "gold":
				userBilling.StripePlanId = "Ultimate"
			}

			userTrialExpires := userBilling.Expires.AddDate(0, 0, -1).Format("2006-01-02")

			data := map[string]interface{}{
				"userBillingTrialExpires": userTrialExpires,
				"userBilling":             userBilling,
				"userEmail":               user.Email,
				"userActive":              user.IsActive,
				csrf.TemplateTag:          csrf.TemplateField(r),
			}

			t := template.New("billing.html")
			t, _ = t.ParseFiles("billing/billing.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func PaymentMethodsPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		userBilling, err := controllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			cards, err := billing.GetUserCards(r, user, &userBilling)
			if err != nil {
				cards = []billing.Card{}
			}

			data := map[string]interface{}{
				"userEmail":      user.Email,
				"userCards":      cards,
				"cardsOnFile":    len(userBilling.CardsOnFile),
				csrf.TemplateTag: csrf.TemplateField(r),
			}

			t := template.New("payments.html")
			t, _ = t.ParseFiles("billing/payments.html")
			t.Execute(w, data)

		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func PaymentMethodsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := controllers.GetCurrentUser(c, r)

		stripeToken := r.FormValue("stripeToken")

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		userBilling, err := controllers.GetUserBilling(c, r, user)
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}

		err = billing.AddPaymentsToCustomer(r, user, &userBilling, stripeToken)

		// Throw error message to user
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "/api/billing/payment-methods", 302)
			return
		}

		http.Redirect(w, r, "/api/billing/payment-methods", 302)
		return
	}
}
