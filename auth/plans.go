package auth

import (
	"net/http"
	"text/template"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

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

		billing, err := GetUserBilling(c, r, user)

		// If the user has already had a trial
		if billing.HasTrial {
			http.Redirect(w, r, "/api/billing/plans", 302)
			return
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://site.newsai.org/", 302)
			return
		}

		data := map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t := template.New("trial.html")
		t, _ = t.ParseFiles("billing/trial.html")
		t.Execute(w, data)
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

		log.Infof(c, "%v", plan)
	}
}
