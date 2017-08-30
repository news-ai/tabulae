package updates

import (
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	apiControllers "github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/controllers"
)

type Social struct {
	Network          string `json:"network"`
	Username         string `json:"username"`
	PrivateOrInvalid string `json:"privateorinvalid"`
}

type SocialToDetails struct {
	Network  string `json:"network"`
	Username string `json:"username"`
	FullName string `json:"fullname"`
}

func SocialUsernameToDetails(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// User has to be logged in
	user, err := apiControllers.GetCurrentUser(c, r)
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
	var socialData SocialToDetails
	err = decoder.Decode(buf, &socialData)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	contacts, err := controllers.FilterContacts(c, r, socialData.Network, socialData.Username)
	if err != nil {
		log.Errorf(c, "%v", socialData)
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	for i := 0; i < len(contacts); i++ {
		// If the contact does not have a first/last name & the full name from the network is not empty
		if contacts[i].FirstName == "" && contacts[i].LastName == "" && socialData.FullName != "" {
			fullNameSplit := strings.Split(socialData.FullName, " ")
			if len(fullNameSplit) > 1 {
				contacts[i].FirstName = fullNameSplit[0]
				contacts[i].LastName = fullNameSplit[1]
			} else {
				contacts[i].FirstName = fullNameSplit[0]
			}
			controllers.Save(c, r, &contacts[i])
		}
	}

	// If successful
	w.WriteHeader(200)
	return
}

func SocialUsernameInvalid(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// User has to be logged in
	user, err := apiControllers.GetCurrentUser(c, r)
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
		log.Errorf(c, "%v", socialData)
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	for i := 0; i < len(contacts); i++ {
		switch socialData.Network {
		case "Twitter":
			switch socialData.PrivateOrInvalid {
			case "Invalid":
				contacts[i].TwitterInvalid = true
			case "Private":
				contacts[i].TwitterPrivate = true
			}
		case "Instagram":
			switch socialData.PrivateOrInvalid {
			case "Invalid":
				contacts[i].InstagramInvalid = true
			case "Private":
				contacts[i].InstagramPrivate = true
			}
		}
		controllers.Save(c, r, &contacts[i])
	}

	// If successful
	w.WriteHeader(200)
	return
}
