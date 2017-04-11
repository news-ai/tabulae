package auth

import (
	"fmt"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/julienschmidt/httprouter"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/oauth2/outlook"
	"github.com/news-ai/web/utilities"

	"golang.org/x/oauth2"
)

var (
	outlookOauthConfig = &oauth2.Config{
		RedirectURL:  "https://tabulae.newsai.org/api/auth/outlookcallback",
		ClientID:     os.Getenv("OUTLOOKAUTHKEY"),
		ClientSecret: os.Getenv("OUTLOOKAUTHSECRET"),
		Scopes: []string{
			"https://outlook.office.com/mail.readwrite",
			"https://outlook.office.com/mail.send",
			"openid",
			"profile",
			"offline_access",
		},
		Endpoint: outlook.Endpoint,
	}
)

// Handler to redirect user to the Outlook OAuth2 page
func OutlookLoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// Make sure the user has been logged in when at outlook auth
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, "user not logged in")
		return
	}

	// Generate a random state that we identify the user with
	state := utilities.RandToken()

	// Save the session for each of the users
	session, err := Store.Get(r, "sess")
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	session.Values["state"] = state
	session.Values["outlook"] = "yes"
	session.Values["outlook_email"] = user.Email

	if r.URL.Query().Get("next") != "" {
		session.Values["next"] = r.URL.Query().Get("next")
	}

	err = session.Save(r, w)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	// Redirect the user to the login page
	url := outlookOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, 302)
}

func OutlookCallbackHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	session, err := Store.Get(r, "sess")
	if err != nil {
		log.Infof(c, "%v", err)
		fmt.Fprintln(w, "aborted")
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		log.Errorf(c, "%v", "no state match; possible csrf OR cookies not enabled")
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}

	tkn, err := outlookOauthConfig.Exchange(c, r.URL.Query().Get("code"))

	if err != nil {
		log.Errorf(c, "%v", "there was an issue getting your token")
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}

	if !tkn.Valid() {
		log.Errorf(c, "%v", "retreived invalid token")
		fmt.Fprintln(w, "retreived invalid token")
		return
	}

	log.Infof(c, "%v", tkn.AccessToken)
	log.Infof(c, "%v", tkn.Expiry)
	log.Infof(c, "%v", tkn.RefreshToken)
	log.Infof(c, "%v", tkn.TokenType)

	// Make sure the user has been logged in when at outlook auth
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, "user not logged in")
		return
	}

	log.Infof(c, "%v", user.Email)

	http.Redirect(w, r, "/", 302)
}
