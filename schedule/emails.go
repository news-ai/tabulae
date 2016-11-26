package schedule

import (
	"net/http"
	// "time"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/controllers"
	"google.golang.org/appengine/log"
	// "github.com/news-ai/tabulae/emails"
	// "github.com/news-ai/tabulae/sync"
	// "github.com/news-ai/web/errors"
)

// When the email is "Delievered == false" and has a "SendAt" date
// And "Cancel == false"

func SchedueleEmailTask(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	emails, err := controllers.GetCurrentSchedueledEmails(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(500)
		return
	}

	log.Infof(c, "%v", emails)

	w.WriteHeader(200)
	return
}
