package updates

import (
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/pquerna/ffjson/ffjson"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"

	"github.com/qedus/nds"

	tabulaeControllers "github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/sync"

	nError "github.com/news-ai/web/errors"
	"github.com/news-ai/web/utilities"
)

type EmailSendUpdate struct {
	EmailId    int64  `json:"emailid"`
	Method     string `json:"method"`
	Delievered bool   `json:"delivered"`

	SendId   string `json:"sendid"`
	ThreadId string `json:"threadid"`
}

func incomingUpdates(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Only listens to POST method
	switch r.Method {
	case "POST":
		buf, _ := ioutil.ReadAll(r.Body)

		decoder := ffjson.NewDecoder()
		var emailSendUpdate []EmailSendUpdate
		err := decoder.Decode(buf, &emailSendUpdate)
		if err != nil {
			log.Errorf(c, "%v", err)
			nError.ReturnError(w, http.StatusInternalServerError, "Updates handing error", err.Error())
			return
		}

		emails := []models.Email{}
		memcacheKeys := []string{}
		emailIds := []int64{}
		keys := []*datastore.Key{}
		for i := 0; i < len(emailSendUpdate); i++ {
			email, _, err := tabulaeControllers.GetEmailByIdUnauthorized(c, r, emailSendUpdate[i].EmailId)
			if err != nil {
				log.Errorf(c, "%v", err)
				continue
			}

			email.IsSent = true
			email.Delievered = emailSendUpdate[i].Delievered
			email.Method = emailSendUpdate[i].Method

			switch emailSendUpdate[i].Method {
			case "sendgrid":
				email.SendGridId = emailSendUpdate[i].SendId
			case "gmail":
				email.GmailId = emailSendUpdate[i].SendId
				email.GmailThreadId = emailSendUpdate[i].ThreadId
			}

			// Invalidate memcache for this particular campaign
			memcacheKey := tabulaeControllers.GetEmailCampaignKey(email)
			memcacheKeys = append(memcacheKeys, memcacheKey)

			keys = append(keys, email.Key(c))
			emailIds = append(emailIds, email.Id)
			emails = append(emails, email)
		}

		ks := []*datastore.Key{}
		err = nds.RunInTransaction(c, func(ctx context.Context) error {
			contextWithTimeout, _ := context.WithTimeout(c, time.Second*150)
			ks, err = nds.PutMulti(contextWithTimeout, keys, emails)
			if err != nil {
				log.Errorf(c, "%v", err)
				return err
			}
			return nil
		}, nil)

		if err != nil {
			log.Errorf(c, "%v", err)
		}

		if len(memcacheKeys) > 0 {
			noDuplicatesMemcache := utilities.RemoveDuplicatesUnordered(memcacheKeys)
			log.Infof(c, "%v", noDuplicatesMemcache)
			err = memcache.DeleteMulti(c, noDuplicatesMemcache)
			if err != nil {
				log.Warningf(c, "%v", err)
			}
		}

		if len(emailIds) > 0 {
			sync.EmailResourceBulkSync(r, emailIds)
		}

		w.WriteHeader(200)
		return
	}

	nError.ReturnError(w, http.StatusInternalServerError, "Updates handing error", "method not implemented")
	return
}

func init() {
	http.HandleFunc("/incoming", internalTrackerHandler)
	http.HandleFunc("/updates", incomingUpdates)
}
