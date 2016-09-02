package notifications

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/channel"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/web/errors"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func GetUserToken(c context.Context, r *http.Request) (interface{}, error) {
	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	token, err := channel.Create(c, currentUser.Id)
	if err != nil {
		log.Errorf(c, "channel.Create: %v", err)
		return nil, err
	}

	tokenResponse := TokenResponse{}
	tokenResponse.Token = token
	return tokenResponse, nil
}
