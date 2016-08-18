package auth

import (
	"errors"
	"net/http"
	"os"

	"github.com/news-ai/gaesessions"
)

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

var Store = gaesessions.NewMemcacheDatastoreStore("", "",
	gaesessions.DefaultNonPersistentSessionDuration,
	[]byte(os.Getenv("SECRETKEY")))

// Gets the email of the current user that is logged in
func GetCurrentUserEmail(r *http.Request) (string, error) {
	session, err := Store.Get(r, "sess")
	if err != nil {
		return "", errors.New("No user logged in")
	}
	if session.Values["email"] == nil {
		return "", errors.New("No user logged in")
	}
	return session.Values["email"].(string), nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "sess")
	delete(session.Values, "state")
	delete(session.Values, "email")
	session.Save(r, w)

	if r.URL.Query().Get("next") != "" {
		http.Redirect(w, r, r.URL.Query().Get("next"), 302)
		return
	}

	http.Redirect(w, r, "/api/auth", 302)
}
