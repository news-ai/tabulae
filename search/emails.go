package search

import (
	"net/http"
	// "net/url"
	"time"
	// "strconv"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticEmailLog        *elastic.Elastic
	elasticEmailTimeseries *elastic.Elastic
	elasticEmails          *elastic.Elastic
)

type Email struct {
	Method string `json:"method"`

	// Which list it belongs to
	ListId     int64 `json:"listid" apiModel:"List"`
	TemplateId int64 `json:"templateid" apiModel:"Template"`
	ContactId  int64 `json:"contactId" apiModel:"Contact"`

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

	SendGridOpened  int `json:"sendgridopened"`
	SendGridClicked int `json:"sendgridclicked"`

	Archived bool `json:"archived"`

	IsSent bool `json:"issent"` // Basically if the user has clicked on "/send"

	Id int64 `json:"id" datastore:"-"`

	Type string `json:"type" datastore:"-"`

	CreatedBy int64 `json:"createdby" apiModel:"User"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func (e *Email) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(e, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchEmail(c context.Context, elasticQuery interface{}) (interface{}, int, error) {
	hits, err := elasticEmailLog.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	log.Infof(c, "%v", hits)

	emailLogHits := []interface{}{}
	for i := 0; i < len(hits.Hits); i++ {
		emailLogHits = append(emailLogHits, hits.Hits[i].Source.Data)
	}

	return emailLogHits, len(emailLogHits), nil
}

func searchEmailTimeseries(c context.Context, elasticQuery interface{}) (interface{}, int, error) {
	hits, err := elasticEmailTimeseries.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	emailTimeseriesHits := []interface{}{}
	for i := 0; i < len(hits.Hits); i++ {
		emailTimeseriesHits = append(emailTimeseriesHits, hits.Hits[i].Source.Data)
	}

	return emailTimeseriesHits, len(emailTimeseriesHits), nil
}

func searchEmailQuery(c context.Context, elasticQuery interface{}) ([]Email, int, error) {
	hits, err := elasticEmails.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Email{}, 0, err
	}

	emailHits := hits.Hits
	emailLogHits := []Email{}
	for i := 0; i < len(emailHits); i++ {
		rawFeed := emailHits[i].Source.Data
		rawMap := rawFeed.(map[string]interface{})
		email := Email{}
		err := email.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		email.Type = "emails"
		emailLogHits = append(emailLogHits, email)
	}

	return emailLogHits, len(emailLogHits), nil
}

func SearchEmailTimeseriesByUserId(c context.Context, r *http.Request, user models.User) (interface{}, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUserIdQuery := ElasticUserIdQuery{}
	elasticUserIdQuery.Term.UserId = user.Id
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUserIdQuery)

	elasticDateQuery := ElasticSortDataQuery{}
	elasticDateQuery.Date.Order = "desc"
	elasticDateQuery.Date.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticDateQuery)

	return searchEmailTimeseries(c, elasticQuery)
}

func SearchEmailLogByEmailId(c context.Context, r *http.Request, user models.User, emailId int64) (interface{}, int, error) {
	if emailId == 0 {
		return nil, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticEmailIdQuery := ElasticEmailIdQuery{}
	elasticEmailIdQuery.Term.EmailId = emailId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmailIdQuery)

	return searchEmail(c, elasticQuery)
}

func SearchEmailsByQuery(c context.Context, r *http.Request, user models.User, searchQuery string) (interface{}, int, error) {
	if searchQuery == "" {
		return nil, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticIsSentQuery := ElasticIsSentQuery{}
	elasticIsSentQuery.Term.IsSent = true

	elasticCancelQuery := ElasticCancelQuery{}
	elasticCancelQuery.Term.Cancel = false

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = searchQuery

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsSentQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCancelQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	elasticCreatedQuery := ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmailQuery(c, elasticQuery)
}

func SearchEmailsByQueryFields(c context.Context, r *http.Request, user models.User, emailDate string, emailSubject string) (interface{}, int, error) {
	if emailDate == "" && emailSubject == "" {
		return nil, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticIsSentQuery := ElasticIsSentQuery{}
	elasticIsSentQuery.Term.IsSent = true

	elasticCancelQuery := ElasticCancelQuery{}
	elasticCancelQuery.Term.Cancel = false

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsSentQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCancelQuery)

	if emailDate != "" {
		elasticCreatedFilterQuery := ElasticCreatedRangeQuery{}
		elasticCreatedFilterQuery.Range.DataCreated.From = emailDate + "T00:00:00"
		elasticCreatedFilterQuery.Range.DataCreated.To = emailDate + "T23:59:59"
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedFilterQuery)
	}

	if emailSubject != "" {
		elasticSubjectQuery := ElasticSubjectQuery{}
		elasticSubjectQuery.Term.Subject = emailSubject
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticSubjectQuery)
	}

	elasticCreatedQuery := ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmailQuery(c, elasticQuery)
}

func SearchEmailsByDateAndSubject(c context.Context, r *http.Request, user models.User, emailDate string, subject string, baseSubject string) ([]Email, int, error) {
	if emailDate == "" {
		return nil, 0, nil
	}

	elasticQuery := elastic.ElasticQueryWithShould{}
	elasticQuery.Size = 10000
	elasticQuery.From = 0

	elasticCreatedByQuery := ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticIsSentQuery := ElasticIsSentQuery{}
	elasticIsSentQuery.Term.IsSent = true

	elasticCancelQuery := ElasticCancelQuery{}
	elasticCancelQuery.Term.Cancel = false

	elasticCreatedFilterQuery := ElasticCreatedRangeQuery{}
	elasticCreatedFilterQuery.Range.DataCreated.From = emailDate + "T00:00:00"
	elasticCreatedFilterQuery.Range.DataCreated.To = emailDate + "T23:59:59"

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	if baseSubject == "" {
		elasticSubjectQuery := ElasticSubjectQuery{}
		elasticSubjectQuery.Term.Subject = subject
		elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Must, elasticSubjectQuery)
	} else {
		elasticBaseSubjectQuery := ElasticBaseSubjectQuery{}
		elasticBaseSubjectQuery.Term.BaseSubject = baseSubject
		elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Must, elasticBaseSubjectQuery)
	}

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsSentQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCancelQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedFilterQuery)

	elasticCreatedQuery := ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmailQuery(c, elasticQuery)
}
