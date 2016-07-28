package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"github.com/news-ai/tabulae/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://tabulae.newsai.org/api/auth/callback",
		ClientID:     os.Getenv("GOOGLEAUTHKEY"),
		ClientSecret: os.Getenv("GOOGLEAUTHSECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/gmail.compose",
			"https://www.googleapis.com/auth/gmail.send",
		},
		Endpoint: google.Endpoint,
	}
)

func SetRedirectURL() {
	googleOauthConfig.RedirectURL = utils.APIURL + "/auth/callback"
}

// Handler to redirect user to the Google OAuth2 page
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a random state that we identify the user with
	state := utils.RandToken()

	// Save the session for each of the users
	session, _ := Store.Get(r, "sess")
	session.Values["state"] = state

	if r.URL.Query().Get("next") != "" {
		session.Values["next"] = r.URL.Query().Get("next")
	}

	session.Save(r, w)

	// Redirect the user to the login page
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

// Handler to get information when callback comes back from Google
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	session, err := Store.Get(r, "sess")
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}

	tkn, err := googleOauthConfig.Exchange(c, r.URL.Query().Get("code"))

	if err != nil {
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}

	if !tkn.Valid() {
		fmt.Fprintln(w, "retreived invalid token")
		return
	}

	client := urlfetch.Client(c)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=" + tkn.AccessToken)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	// Decode JSON from Google
	decoder := json.NewDecoder(resp.Body)
	var googleUser User
	err = decoder.Decode(&googleUser)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	session.Values["email"] = googleUser.Email
	session.Values["id"] = googleUser.ID
	session.Values["verifiedemail"] = googleUser.VerifiedEmail
	session.Values["name"] = googleUser.Name
	session.Values["given_name"] = googleUser.GivenName
	session.Values["family_name"] = googleUser.FamilyName
	session.Values["picture"] = googleUser.Picture
	session.Values["locale"] = googleUser.Locale
	session.Values["hd"] = googleUser.Hd
	session.Values["accessToken"] = tkn.AccessToken
	session.Save(r, w)

	if session.Values["next"] != nil {
		http.Redirect(w, r, session.Values["next"].(string), 302)
	}

	http.Redirect(w, r, "/", 302)
}
