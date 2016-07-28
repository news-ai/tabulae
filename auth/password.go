package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"appengine"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

func PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Setup to authenticate the user into the API
	email := r.FormValue("email")
	password := r.FormValue("password")

	a := models.User{}
	c.Infof("%v", a, email, password)

	// Generate a random state that we identify the user with
	state := utils.RandToken()

	// Save the session for each of the users
	session, _ := Store.Get(r, "sess")
	session.Values["state"] = state
	session.Save(r, w)

	session.Values["email"] = email
}

func PasswordLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	t := template.New("Login template")

	cwd, _ := os.Getwd()

	c.Infof("%v", filepath.Join(cwd, "../auth/login.html"))

	t, _ = t.ParseFiles(filepath.Join(cwd, "../auth/login.html"))
	t.Execute(w, "")
}
