package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	"github.com/qedus/nds"
)

type EmailProviderLimits struct {
	SendGrid       int `json:"sendgrid"`
	SendGridLimits int `json:"sendgridLimits"`
	Outlook        int `json:"outlook"`
	OutlookLimits  int `json:"outlookLimits"`
	Gmail          int `json:"gmail"`
	GmailLimits    int `json:"gmailLimits"`
	SMTP           int `json:"smtp"`
	SMTPLimits     int `json:"smtpLimits"`
}

type BulkSendEmailIds struct {
	EmailIds []int64 `json:"emailids"`
}

type SMTPSettings struct {
	Servername string `json:"servername"`

	EmailUser     string `json:"emailuser"`
	EmailPassword string `json:"emailpassword"`
}

type SMTPEmailSettings struct {
	Servername string `json:"servername"`

	EmailUser     string `json:"emailuser"`
	EmailPassword string `json:"emailpassword"`

	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type UserEmailSetting struct {
	SMTPUsername string `json:"smtpusername"`
	SMTPPassword string `json:"smtppassword"`
}

type EmailSetting struct {
	apiModels.Base

	SMTPServer  string `json:"SMTPServer"`
	SMTPPortTLS int    `json:"SMTPPortTLS"`
	SMTPPortSSL int    `json:"SMTPPortSSL"`
	SMTPSSLTLS  bool   `json:"SMTPSSLTLS"`

	IMAPServer  string `json:"IMAPServer"`
	IMAPPortTLS int    `json:"IMAPPortTLS"`
	IMAPPortSSL int    `json:"IMAPPortSSL"`
	IMAPSSLTLS  bool   `json:"IMAPSSLTLS"`
}

type Email struct {
	apiModels.Base

	Method string `json:"method"`

	// Which list it belongs to
	ListId     int64 `json:"listid" apiModel:"List"`
	TemplateId int64 `json:"templateid" apiModel:"Template"`
	ContactId  int64 `json:"contactId" apiModel:"Contact"`
	ClientId   int64 `json:"clientid"`

	FromEmail string `json:"fromemail"`

	Sender      string `json:"sender"`
	To          string `json:"to"`
	Subject     string `json:"subject" datastore:",noindex"`
	BaseSubject string `json:"baseSubject" datastore:",noindex"`
	Body        string `json:"body" datastore:",noindex"`

	CC  []string `json:"cc"`  // Carbon copy email addresses
	BCC []string `json:"bcc"` // Blind carbon copy email addresses

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	SendAt time.Time `json:"sendat"`

	SendGridId  string `json:"-"`
	SparkPostId string `json:"-"`
	BatchId     string `json:"batchid"`

	GmailId       string `json:"gmailid"`
	GmailThreadId string `json:"gmailthreadid"`

	TeamId int64 `json:"teamid"`

	Attachments []int64 `json:"attachments" datastore:",noindex" apiModel:"File"`

	Delievered    bool   `json:"delivered"` // The email has been officially sent by our platform
	BouncedReason string `json:"bouncedreason"`
	Bounced       bool   `json:"bounced"`
	Clicked       int    `json:"clicked"`
	Opened        int    `json:"opened"`
	Spam          bool   `json:"spam"`
	Cancel        bool   `json:"cancel"`
	Dropped       bool   `json:"dropped"`

	SendGridOpened  int `json:"sendgridopened"`
	SendGridClicked int `json:"sendgridclicked"`

	Archived bool `json:"archived"`

	IsSent bool `json:"issent"` // Basically if the user has clicked on "/send"
}

/*
* Public methods
 */

/*
* Create methods
 */

func (e *Email) Key(c context.Context) *datastore.Key {
	return e.BaseKey(c, "Email")
}

func (e *Email) Create(c context.Context, r *http.Request, currentUser apiModels.User) (*Email, error) {
	e.IsSent = false
	e.CreatedBy = currentUser.Id
	e.Created = time.Now()

	_, err := e.Save(c)
	return e, err
}

func (es *EmailSetting) Create(c context.Context, r *http.Request, currentUser apiModels.User) (*EmailSetting, error) {
	es.CreatedBy = currentUser.Id
	es.Created = time.Now()

	_, err := es.Save(c)
	return es, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (e *Email) Save(c context.Context) (*Email, error) {
	// Update the Updated time
	e.Updated = time.Now()

	k, err := nds.Put(c, e.BaseKey(c, "Email"), e)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	e.Id = k.IntID()
	return e, nil
}

// Function to save a new email into App Engine
func (es *EmailSetting) Save(c context.Context) (*EmailSetting, error) {
	// Update the Updated time
	es.Updated = time.Now()

	k, err := nds.Put(c, es.BaseKey(c, "EmailSetting"), es)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	es.Id = k.IntID()
	return es, nil
}

func (e *Email) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(e, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
