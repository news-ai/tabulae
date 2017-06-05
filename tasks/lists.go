package tasks

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	apiControllers "github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/web/errors"
)

func ListsToIncludeTeamId(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	allMediaLists, err := controllers.GetAllMediaLists(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get media lists", err.Error())
		return
	}

	for i := 0; i < len(allMediaLists); i++ {
		mediaListUser, err := apiControllers.GetUserByIdUnauthorized(c, r, allMediaLists[i].CreatedBy)
		if err != nil {
			log.Errorf(c, "%v", err)
			errors.ReturnError(w, http.StatusInternalServerError, "Could not get user", err.Error())
			return
		}

		if mediaListUser.TeamId != 0 {
			allMediaLists[i].TeamId = mediaListUser.TeamId
			allMediaLists[i].Save(c)
		}
	}

	// If successful
	w.WriteHeader(200)
	return
}
