package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"github.com/news-ai/tabulae/utils"

	"github.com/hnakamur/gaesessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "/auth/callback",
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

var store = gaesessions.NewMemcacheDatastoreStore("", "",
	gaesessions.DefaultNonPersistentSessionDuration,
	[]byte("Cab7MNoPdBdX%fxN?yg3yWVM4^7KecETjfem6HwXizQqZTsG4#"))

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	Hd            string `json:"hd"`
}

func SetRedirectURL() {
	googleOauthConfig.RedirectURL = utils.APIURL + googleOauthConfig.RedirectURL
}

// State can be some kind of random generated hash string.
// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Gets the email of the current user that is logged in
func GetCurrentUserEmail(r *http.Request) (string, error) {
	session, err := store.Get(r, "sess")
	if err != nil {
		return "", errors.New("No user logged in")
	}
	if session.Values["email"] == nil {
		return "", errors.New("No user logged in")
	}
	return session.Values["email"].(string), nil
}

// Gets the full details of the current user
func GetUserDetails(r *http.Request) (map[string]string, error) {
	session, err := store.Get(r, "sess")

	// If there is no session
	if err != nil {
		return nil, errors.New("No user logged in")
	}

	// If there exists no email then user not logged in
	if session.Values["email"].(string) == "" {
		return nil, errors.New("No user logged in")
	}

	// Takes interface{} values and converts them into string
	userDetails := map[string]string{}
	for k, v := range session.Values {
		key := fmt.Sprint(k)
		value := fmt.Sprint(v)
		userDetails[key] = value
	}

	return userDetails, nil
}

// Handler to redirect user to the Google OAuth2 page
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a random state that we identify the user with
	state := randToken()

	// Save the session for each of the users
	session, _ := store.Get(r, "sess")
	session.Values["state"] = state
	session.Save(r, w)

	// Redirect the user to the login page
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

// Handler to get information when callback comes back from Google
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	session, err := store.Get(r, "sess")
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

	http.Redirect(w, r, "/", 302)
}
