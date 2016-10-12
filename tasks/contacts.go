package tasks

import (
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"
)

type Social struct {
	Network  string `json:"network"`
	Username string `json:"username"`
}

func SocialUsernameInvalid(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// User has to be logged in
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	// User has to be an admin
	if !user.IsAdmin {
		log.Errorf(c, "%v", "User that hit the social username invalid method is not an admin")
		w.WriteHeader(500)
		return
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var socialData Social
	err = decoder.Decode(buf, &socialData)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	contacts, err := controllers.FilterContacts(c, r, socialData.Network, socialData.Username)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	for i := 0; i < len(contacts); i++ {
		switch socialData.Network {
		case "Twitter":
			contacts[i].TwitterInvalid = true
		case "Instagram":
			contacts[i].InstagramInvalid = true
		}
		controllers.Save(c, r, &contacts[i])
	}

	// If successful
	w.WriteHeader(200)
	return
}
