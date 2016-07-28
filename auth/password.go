package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"appengine"

	"github.com/news-ai/tabulae/models"
	// "github.com/news-ai/tabulae/utils"
	// "github.com/gorilla/csrf"
)

func PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Setup to authenticate the user into the API
	email := r.FormValue("email")
	password := r.FormValue("password")

	a := models.User{}
	c.Infof("%v", a, email, password)

	// // Generate a random state that we identify the user with
	// state := utils.RandToken()

	// // Save the session for each of the users
	// session, _ := Store.Get(r, "sess")
	// session.Values["state"] = state
	// session.Save(r, w)

	// //

	// // Now that the user is created/retrieved save the email in the session
	// session.Values["email"] = email
	// session.Save(r, w)
}

// Don't start their session here, but when they login to the platform.
// This is just to give them the ability to register an account.
func PasswordRegisterHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Setup to authenticate the user into the API
	firstName := r.FormValue("firstname")
	lastName := r.FormValue("lastname")
	email := r.FormValue("email")
	password := r.FormValue("password")

	a := models.User{}
	c.Infof("%v", a, firstName, lastName, email, password)

	// Hash the password and save it into the datastore

	// Send an email confirmation
}

// Takes ?next as well. Create a session for the person.
// Will post data to the password login handler.
// Redirect to the ?next parameter.
// Put CSRF token into the login handler.
func PasswordLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	cwd, _ := os.Getwd()

	if r.URL.Query().Get("next") != "" {
		session, _ := Store.Get(r, "sess")
		session.Values["next"] = r.URL.Query().Get("next")
		session.Save(r, w)
	}

	file := filepath.Join(cwd, "../auth/static/login.html")
	t := template.New("login.html")
	t, _ = t.ParseFiles(file)
	t.Execute(w, "")
}

func PasswordRegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	cwd, _ := os.Getwd()

	if r.URL.Query().Get("next") != "" {
		session, _ := Store.Get(r, "sess")
		session.Values["next"] = r.URL.Query().Get("next")
		session.Save(r, w)
	}

	file := filepath.Join(cwd, "../auth/static/register.html")
	t := template.New("register.html")
	t, _ = t.ParseFiles(file)
	t.Execute(w, "")
}
