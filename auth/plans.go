package auth

import (
	"net/http"
	"text/template"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/controllers"

	"github.com/gorilla/csrf"
)

func ChoosePlanHandler() http.HandlerFunc {
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
