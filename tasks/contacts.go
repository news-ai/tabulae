package tasks

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"

	"github.com/news-ai/web/errors"
)

type Social struct {
	Network  string `json:"network"`
	Username string `json:"username"`
}

func SocialUsernameInvalid(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// User has to be logged in
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	// User has to be an admin
	if !user.IsAdmin {
		w.WriteHeader(500)
		return
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var socialData Social
	err = decoder.Decode(buf, &socialData)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	contacts, err := controllers.FilterContact(c, r, socialData.Network, socialData.Username)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	for i := 0; i < len(contacts); i++ {
		controllers.Save(c, r, &contacts[i])
	}

	// If successful
	w.WriteHeader(200)
	return
}
