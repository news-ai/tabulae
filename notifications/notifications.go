package notifications

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/channel"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/web/errors"
)

type ToeknResponse struct {
	Token string `json:"token"`
}

func func_name(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	c := appengine.NewContext(r)

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusUnauthorized, "Authentication Required", "Could not generate token")
        return
	}

	token, err := channel.Create(c, u.ID+key)
	if err != nil {
		log.Errorf(c, "channel.Create: %v", err)
        errors.ReturnError(w, http.StatusUnauthorized, "Token error", "Could not generate token")
		return
	}

    currentUser.

}
