package auth

import (
	"net/http"
	"net/url"
	"text/template"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
	// "github.com/gorilla/csrf"
)

func PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
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
	if isOk {
		// // Now that the user is created/retrieved save the email in the session
		session.Values["email"] = email
		session.Save(r, w)

		if session.Values["next"] != nil {
			http.Redirect(w, r, session.Values["next"].(string), 302)
		}
	}
	wrongPasswordMessage := "You entered the wrong password!"
	http.Redirect(w, r, url.QueryEscape("/api/auth?success=false&message="+wrongPasswordMessage), 302)
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
	isOk, err := models.RegisterUser(r, user)

	if !isOk && err != nil {
		// Redirect user back to login page
		http.Redirect(w, r, url.QueryEscape("/api/auth?success=false&message="+err.Error()), 302)
	}

	// Send an email confirmation
	emailConfirmation := emails.Email{}
	emailConfirmation.To = []string{email}
	emailConfirmation.FirstName = firstName
	emails.SendConfirmationEmail(r, emailConfirmation)

	// Redirect user back to login page
	confirmationMessage := "We sent you a confirmation email!"
	http.Redirect(w, r, url.QueryEscape("/api/auth?success=true&message="+confirmationMessage), 302)
}

// Takes ?next as well. Create a session for the person.
// Will post data to the password login handler.
// Redirect to the ?next parameter.
// Put CSRF token into the login handler.
func PasswordLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.URL.Query().Get("next") != "" {
		session, _ := Store.Get(r, "sess")
		session.Values["next"] = r.URL.Query().Get("next")
		session.Save(r, w)
	}

	t := template.New("login.html")
	t, err := t.ParseFiles("auth/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, "")
	if err != nil {
		log.Errorf(c, "could not put into datastore: %v", err)
	}
}

func PasswordRegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.URL.Query().Get("next") != "" {
		session, _ := Store.Get(r, "sess")
		session.Values["next"] = r.URL.Query().Get("next")
		session.Save(r, w)
	}

	t := template.New("register.html")
	t, _ = t.ParseFiles("auth/register.html")
	err := t.Execute(w, "")
	if err != nil {
		log.Errorf(c, "could not put into datastore: %v", err)
	}
}
