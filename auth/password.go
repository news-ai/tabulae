package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"appengine"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
	// "github.com/gorilla/csrf"
)

func PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Setup to authenticate the user into the API
	email := r.FormValue("email")
	password := r.FormValue("password")

	// // Generate a random state that we identify the user with
	state := utils.RandToken()

	// // Save the session for each of the users
	session, _ := Store.Get(r, "sess")
	session.Values["state"] = state
	session.Save(r, w)

	isOk, _ := models.ValidateUserPassword(r, email, password)
	c.Infof("%v", isOk)
	if isOk {
		// // Now that the user is created/retrieved save the email in the session
		session.Values["email"] = email
		session.Save(r, w)

		if session.Values["next"] != nil {
			http.Redirect(w, r, session.Values["next"].(string), 302)
		}
	}
	http.Redirect(w, r, "/", 302)
}

// Don't start their session here, but when they login to the platform.
// This is just to give them the ability to register an account.
func PasswordRegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Setup to authenticate the user into the API
	firstName := r.FormValue("firstname")
	lastName := r.FormValue("lastname")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Hash the password and save it into the datastore
	hashedPassword, _ := utils.HashPassword(password)

	user := models.User{}
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = hashedPassword
	user.EmailConfirmed = false

	// Register user
	models.RegisterUser(r, user)

	// Send an email confirmation

	// Redirect user back to login page
	http.Redirect(w, r, "/api/auth?success=true", 302)
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
