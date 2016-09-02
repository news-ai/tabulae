package notifications

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine/channel"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func UserConnect(w http.ResponseWriter, r *http.Request) {
	// clientID := r.FormValue("from")
}

func UserDisconnect(w http.ResponseWriter, r *http.Request) {
	// clientID := r.FormValue("from")
}

func GetUserToken(c context.Context, r *http.Request) (interface{}, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	token, err := channel.Create(c, strconv.FormatInt(currentUser.Id, 10))
	if err != nil {
		log.Errorf(c, "channel.Create: %v", err)
		return nil, err
	}

	tokenResponse := TokenResponse{}
	tokenResponse.Token = token
	return tokenResponse, nil
}
