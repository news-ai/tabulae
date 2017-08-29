package updates

import (
	"io/ioutil"
	"net/http"

	"github.com/pquerna/ffjson/ffjson"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	tabulaeControllers "github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/sync"

	nError "github.com/news-ai/web/errors"
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

		emailIds := []int64{}
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

			email.Save(c)
			emailIds = append(emailIds, email.Id)
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
	http.HandleFunc("/updates", incomingUpdates)
}
