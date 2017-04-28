package search

import (
	"net/http"
	// "net/url"
	// "time"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticEmailCampaign *elastic.Elastic
)

type EmailCampaignResponse struct {
	Date        string `json:"date"`
	Subject     string `json:"subject"`
	UserId      string `json:"userid"`
	BaseSubject string `json:"baseSubject"`

	Delivered              int     `json:"delivered"`
	Opens                  int     `json:"opens"`
	UniqueOpens            int     `json:"uniqueOpens"`
	UniqueOpensPercentage  float32 `json:"uniqueOpensPercentage"`
	Clicks                 int     `json:"clicks"`
	UniqueClicks           int     `json:"uniqueClicks"`
	UniqueClicksPercentage float32 `json:"uniqueClicksPercentage"`
	Bounces                int     `json:"bounces"`
}

type EmailCampaignRequest struct {
	Date        string `json:"date"`
	Subject     string `json:"subject"`
	UserId      string `json:"userid"`
	BaseSubject string `json:"baseSubject"`
}

func (ec *EmailCampaignRequest) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(ec, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchEmailCampaigns(c context.Context, r *http.Request, elasticQuery interface{}, user models.User) (interface{}, int, error) {
	hits, err := elasticEmailCampaign.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	// Get all email campaigns
	emailCampaigns := []EmailCampaignRequest{}
	for i := 0; i < len(hits.Hits); i++ {
		rawEmailCampaign := hits.Hits[i].Source.Data
		rawMap := rawEmailCampaign.(map[string]interface{})
		emailCampaign := EmailCampaignRequest{}
		err := emailCampaign.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
			continue
		}

		emailCampaigns = append(emailCampaigns, emailCampaign)
	}

	// Get all emails for each of the campaigns
	emailCampaignsResponse := []EmailCampaignResponse{}
	for i := 0; i < len(emailCampaigns); i++ {
		emails, _, err := SearchEmailsByDateAndSubject(c, r, user, emailCampaigns[i].Date, emailCampaigns[i].Subject, emailCampaigns[i].BaseSubject)
		if err != nil {
			log.Errorf(c, "%v", err)
			continue
		}

		emailCampaign := EmailCampaignResponse{}
		emailCampaign.Date = emailCampaigns[i].Date
		emailCampaign.UserId = emailCampaigns[i].UserId
		emailCampaign.Subject = emailCampaigns[i].Subject
		emailCampaign.BaseSubject = emailCampaigns[i].BaseSubject

		for x := 0; x < len(emails); x++ {
			if emails[x].Subject == emailCampaign.Subject {
				if emailCampaign.Subject == "" {
					emailCampaign.Subject = "(no subject)"
				}

				emailCampaign.Delivered += 1
				emailCampaign.Opens += emails[x].Opened
				emailCampaign.Clicks += emails[x].Clicked

				if emails[x].Opened > 0 {
					emailCampaign.UniqueOpens += 1
				}

				if emails[x].Clicked > 0 {
					emailCampaign.UniqueClicks += 1
				}

				if emails[x].Bounced {
					emailCampaign.Bounces += 1
				}
			}
		}

		if emailCampaign.Delivered > 0 {
			emailCampaign.UniqueOpensPercentage = 100 * float32(float32(emailCampaign.UniqueOpens)/float32(emailCampaign.Delivered))
			emailCampaign.UniqueClicksPercentage = 100 * float32(float32(emailCampaign.UniqueClicks)/float32(emailCampaign.Delivered))
		}

		emailCampaignsResponse = append(emailCampaignsResponse, emailCampaign)
	}

	return emailCampaignsResponse, len(emailCampaignsResponse), nil
}

func SearchEmailCampaignsByDate(c context.Context, r *http.Request, user models.User) (interface{}, int, error) {
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

	return searchEmailCampaigns(c, r, elasticQuery, user)
}