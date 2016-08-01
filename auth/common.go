package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/hnakamur/gaesessions"
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

// Gets the full details of the current user
func GetUserDetails(r *http.Request) (map[string]string, error) {
	session, err := Store.Get(r, "sess")

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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "sess")
	delete(session.Values, "state")
	delete(session.Values, "email")
	session.Save(r, w)

	http.Redirect(w, r, "/api/auth", 302)
}
