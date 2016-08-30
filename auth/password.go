package auth

import (
	"net/http"
	"net/mail"
	"net/url"
	"text/template"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"

	"github.com/gorilla/csrf"
)

func PasswordLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		// Setup to authenticate the user into the API
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validate email
		validEmail, err := mail.ParseAddress(email)
		if err != nil {
			invalidEmailAlert := url.QueryEscape("The email you entered is not valid!")
			http.Redirect(w, r, "/api/auth?success=false&message="+invalidEmailAlert, 302)
		}

		// // Generate a random state that we identify the user with
		state := utils.RandToken()

		// // Save the session for each of the users
		session, _ := Store.Get(r, "sess")
		session.Values["state"] = state
		session.Save(r, w)

		user, isOk, _ := controllers.ValidateUserPassword(r, validEmail.Address, password)
		if isOk {
			// // Now that the user is created/retrieved save the email in the session
			if !user.EmailConfirmed {
				emailNotConfirmedMessage := url.QueryEscape("You have not confirmed your email yet! Please check your email.")
				http.Redirect(w, r, "/api/auth?success=false&message="+emailNotConfirmedMessage, 302)
				return
			}

			session.Values["email"] = validEmail.Address
			session.Save(r, w)

			if session.Values["next"] != nil {
				returnURL := session.Values["next"].(string)
				u, err := url.Parse(returnURL)
				if err != nil {
					log.Errorf(c, "%v", err)
					http.Redirect(w, r, returnURL, 302)
					return
				}
				if user.LastLoggedIn.IsZero() {
					q := u.Query()
					q.Set("firstTimeUser", "true")
					u.RawQuery = q.Encode()
				}
				http.Redirect(w, r, u.String(), 302)
				return
			}

			http.Redirect(w, r, "/", 302)
			return
		}
		wrongPasswordMessage := url.QueryEscape("You entered the wrong password!")
		http.Redirect(w, r, "/api/auth?success=false&message="+wrongPasswordMessage, 302)
		return
	}
}

func ForgetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		// Forget password
		email := r.FormValue("email")

		// Validate email
		validEmail, err := mail.ParseAddress(email)
		if err != nil {
			invalidEmailAlert := url.QueryEscape("The email you entered is not valid!")
			http.Redirect(w, r, "/api/auth?success=false&message="+invalidEmailAlert, 302)
		}

		user, err := controllers.GetUserByEmail(c, email)
		if err != nil {
			noUserErr := url.QueryEscape("There is no user with this email!")
			http.Redirect(w, r, "/api/auth?success=false&message="+noUserErr, 302)
		}

		if user.GoogleId != "" {
			googleAuthErr := url.QueryEscape("You signed up with Google Authentication!")
			http.Redirect(w, r, "/api/auth?success=false&message="+googleAuthErr, 302)
		}

		user.ResetPasswordCode = utils.RandToken()
		user.Save(c)

		emailReset, _ := controllers.CreateEmailInternal(r, validEmail.Address, user.FirstName, user.LastName)
		emailSent, emailId, err := emails.SendResetEmail(r, emailReset, user.ResetPasswordCode)
		if !emailSent || err != nil {
			// Redirect user back to login page
			log.Errorf(c, "%v", "Reset email was not sent for "+email)
			log.Errorf(c, "%v", err)
			emailResetErr := url.QueryEscape("Could not send a reset email. We'll fix this soon!")
			http.Redirect(w, r, "/api/auth?success=false&message="+emailResetErr, 302)
			return
		}

		emailReset.MarkSent(c, emailId)

		// Redirect user back to login page
		resetMessage := url.QueryEscape("We sent you a password reset email!")
		http.Redirect(w, r, "/api/auth?success=true&message="+resetMessage, 302)
	}
}

// Don't start their session here, but when they login to the platform.
// This is just to give them the ability to register an account.
func PasswordRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		// Setup to authenticate the user into the API
		firstName := r.FormValue("firstname")
		lastName := r.FormValue("lastname")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validate email
		validEmail, err := mail.ParseAddress(email)
		if err != nil || email == "" {
			invalidEmailAlert := url.QueryEscape("Validation failed on registration. Sorry about that!")
			http.Redirect(w, r, "/api/auth?success=false&message="+invalidEmailAlert, 302)
		}

		// Hash the password and save it into the datastore
		hashedPassword, _ := utils.HashPassword(password)

		user := models.User{}
		user.FirstName = firstName
		user.LastName = lastName
		user.Email = validEmail.Address
		user.Password = hashedPassword
		user.EmailConfirmed = false
		user.AgreeTermsAndConditions = true
		user.ConfirmationCode = utils.RandToken()

		// Register user
		_, isOk, err := controllers.RegisterUser(r, user)

		if !isOk && err != nil {
			// Redirect user back to login page
			emailRegistered := url.QueryEscape("Email has already been registered")
			http.Redirect(w, r, "/api/auth?success=false&message="+emailRegistered, 302)
			return
		}

		// Email could fail to send if there is no singleUser. Create check later.
		emailConfirmation, _ := controllers.CreateEmailInternal(r, email, firstName, lastName)
		emailSent, emailId, err := emails.SendConfirmationEmail(r, emailConfirmation, user.ConfirmationCode)
		if !emailSent || err != nil {
			// Redirect user back to login page
			log.Errorf(c, "%v", "Confirmation email was not sent for "+email)
			log.Errorf(c, "%v", err)
			emailRegistered := url.QueryEscape("Could not send confirmation email. We'll fix this soon!")
			http.Redirect(w, r, "/api/auth?success=false&message="+emailRegistered, 302)
			return
		}

		emailConfirmation.MarkSent(c, emailId)

		// Redirect user back to login page
		confirmationMessage := url.QueryEscape("We sent you a confirmation email!")
		http.Redirect(w, r, "/api/auth?success=true&message="+confirmationMessage, 302)
	}
}

// Takes ?next as well. Create a session for the person.
// Will post data to the password login handler.
// Redirect to the ?next parameter.
// Put CSRF token into the login handler.
func PasswordLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		_, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has been logged in
			if err == nil {
				http.Redirect(w, r, session.Values["next"].(string), 302)
				return
			}
		}

		// If there is no next and the user is logged in
		if err == nil {
			http.Redirect(w, r, "/", 302)
			return
		}

		// If there is no user then we redirect them to the login page
		t := template.New("login.html")
		t, err = t.ParseFiles("auth/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t.Execute(w, data)
	}
}

// You have to be logged out in order to register a new user
func PasswordRegisterPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		_, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has been logged in
			if err == nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is logged in
		if err == nil {
			http.Redirect(w, r, "/", 302)
			return
		}

		data := map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t := template.New("register.html")
		t, _ = t.ParseFiles("auth/register.html")
		t.Execute(w, data)
	}
}

func ForgetPasswordPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		_, err := controllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has been logged in
			if err == nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is logged in
		if err == nil {
			http.Redirect(w, r, "/", 302)
			return
		}

		data := map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t := template.New("forget.html")
		t, _ = t.ParseFiles("auth/forget.html")
		t.Execute(w, data)
	}
}

func ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		_, err := controllers.GetCurrentUser(c, r)

		// Invalid confirmation message
		invalidResetCode := url.QueryEscape("Your reset code is invalid!")

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has been logged in
			if err == nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is logged in
		if err == nil {
			http.Redirect(w, r, "/", 302)
			return
		}

		// Validate token
		if val, ok := r.URL.Query()["code"]; ok {
			code := val[0]
			codeUnscape, err := url.QueryUnescape(code)
			if err != nil {
				log.Infof(c, "%v", codeUnscape)
				log.Infof(c, "%v", err)
				http.Redirect(w, r, "/api/auth?success=false&message="+invalidResetCode, 302)
				return
			}
			user, err := controllers.GetUserByResetCode(c, codeUnscape)
			if err != nil {
				log.Infof(c, "%v", codeUnscape)
				log.Infof(c, "%v", err)
				http.Redirect(w, r, "/api/auth?success=false&message="+invalidResetCode, 302)
				return
			}
		}

		data := map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t := template.New("reset.html")
		t, _ = t.ParseFiles("auth/reset.html")
		t.Execute(w, data)
	}
}

func EmailConfirmationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)

		// Invalid confirmation message
		invalidConfirmation := url.QueryEscape("Your confirmation code is invalid!")

		if val, ok := r.URL.Query()["code"]; ok {
			code := val[0]
			codeUnscape, err := url.QueryUnescape(code)
			if err != nil {
				log.Infof(c, "%v", codeUnscape)
				log.Infof(c, "%v", err)
				http.Redirect(w, r, "/api/auth?success=false&message="+invalidConfirmation, 302)
			}
			user, err := controllers.GetUserByConfirmationCode(c, codeUnscape)
			if err != nil {
				log.Infof(c, "%v", codeUnscape)
				log.Infof(c, "%v", err)
				http.Redirect(w, r, "/api/auth?success=false&message="+invalidConfirmation, 302)
			}

			_, err = user.ConfirmEmail(c)
			if err != nil {
				log.Infof(c, "%v", err)
				http.Redirect(w, r, "/api/auth?success=false&message="+invalidConfirmation, 302)
			}

			validConfirmation := "Your email has been confirmed. Please proceed to logging in!"
			http.Redirect(w, r, "/api/auth?success=true&message="+validConfirmation, 302)
		}

		http.Redirect(w, r, "/api/auth?success=false&message="+invalidConfirmation, 302)
	}
}
