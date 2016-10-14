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

type FeedUrl struct {
	Url string `json:"url"`
}

func FeedInvalid(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	var url FeedUrl
	err = decoder.Decode(buf, &url)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	feeds, err := controllers.FilterFeeds(c, r, "FeedURL", url.Url)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].ValidFeed = true
		feeds[i].Save(c)
	}

	// If successful
	w.WriteHeader(200)
	return
}
