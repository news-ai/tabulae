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
				// If the user has already had a trial and has expired
				if userBilling.HasTrial && !userBilling.Expires.IsZero() && userBilling.Expires.After(time.Now()) {
					http.Redirect(w, r, "/api/billing/plans", 302)
					return
				}

				// If the user has already had a trial but it has not expired
				if userBilling.HasTrial && !userBilling.Expires.IsZero() && userBilling.Expires.Before(time.Now()) {
					http.Redirect(w, r, "https://site.newsai.org/", 302)
					return
				}
			}

			data := map[string]interface{}{
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
		_, err := controllers.GetCurrentUser(c, r)

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
		_, err := controllers.GetCurrentUser(c, r)

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
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t := template.New("billing.html")
		t, _ = t.ParseFiles("billing/billing.html")
		t.Execute(w, data)
	}
}
