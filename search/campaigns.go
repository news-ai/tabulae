package search

import (
	"net/http"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	"github.com/news-ai/api/models"
	apiSearch "github.com/news-ai/api/search"
	tabulaeModels "github.com/news-ai/tabulae/models"

	elastic "github.com/news-ai/elastic-appengine"
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

	IsScheduled bool `json:"isscheduled"`
	Show        bool `json:"show"`
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

func searchEmailCampaigns(c context.Context, r *http.Request, elasticQuery interface{}, user models.User) (interface{}, int, int, error) {
	hits, err := elasticEmailCampaign.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
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
		limit := 750
		emails, _, total, err := SearchEmailsByDateAndSubject(c, r, user, emailCampaigns[i].Date, emailCampaigns[i].Subject, emailCampaigns[i].BaseSubject, 0, limit)

		if err != nil {
			log.Errorf(c, "%v", err)
			continue
		}

		if emailCampaigns[i].BaseSubject != "" {
			log.Infof(c, "%v", len(emails))
		}

		emailCampaign := EmailCampaignResponse{}
		emailCampaign.Date = emailCampaigns[i].Date
		emailCampaign.UserId = emailCampaigns[i].UserId
		emailCampaign.Subject = emailCampaigns[i].Subject
		emailCampaign.BaseSubject = emailCampaigns[i].BaseSubject

		for x := 0; x < len(emails); x++ {
			emailSubject := emails[x].Subject
			if emails[x].BaseSubject != "" {
				emailSubject = emails[x].BaseSubject
			}
			if emailSubject == emailCampaign.Subject && !emails[x].Archived {
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

		emails = []tabulaeModels.Email{}

		// If we have to loop through
		if total > limit {
			loops := int(float64(total) / float64(limit))
			for x := 1; x < loops; x++ {
				additionalEmails, _, _, err := SearchEmailsByDateAndSubject(c, r, user, emailCampaigns[i].Date, emailCampaigns[i].Subject, emailCampaigns[i].BaseSubject, limit*x, limit)
				if err != nil {
					log.Errorf(c, "%v", err)
					continue
				}

				for y := 0; y < len(additionalEmails); y++ {
					emailSubject := additionalEmails[y].Subject
					if additionalEmails[y].BaseSubject != "" {
						emailSubject = additionalEmails[y].BaseSubject
					}
					if emailSubject == emailCampaign.Subject && !additionalEmails[y].Archived {
						if emailCampaign.Subject == "" {
							emailCampaign.Subject = "(no subject)"
						}

						emailCampaign.Delivered += 1
						emailCampaign.Opens += additionalEmails[y].Opened
						emailCampaign.Clicks += additionalEmails[y].Clicked

						if additionalEmails[y].Opened > 0 {
							emailCampaign.UniqueOpens += 1
						}

						if additionalEmails[y].Clicked > 0 {
							emailCampaign.UniqueClicks += 1
						}

						if additionalEmails[y].Bounced {
							emailCampaign.Bounces += 1
						}
					}
				}
				additionalEmails = []tabulaeModels.Email{}
			}
		}

		if emailCampaign.Delivered > 0 {
			deliveredNumber := emailCampaign.Delivered - emailCampaign.Bounces
			log.Infof(c, "%v", deliveredNumber)
			if deliveredNumber > 0 {
				// For some reason if more people opened it then the number of
				// emails that were delivered then we set a ceiling of 100%
				if emailCampaign.UniqueOpens > deliveredNumber {
					emailCampaign.UniqueOpens = deliveredNumber
				}

				emailCampaign.UniqueOpensPercentage = 100 * float32(float32(emailCampaign.UniqueOpens)/float32(deliveredNumber))
				emailCampaign.UniqueClicksPercentage = 100 * float32(float32(emailCampaign.UniqueClicks)/float32(deliveredNumber))
				emailCampaign.Show = true
			}
		}

		emailCampaignsResponse = append(emailCampaignsResponse, emailCampaign)
	}

	return emailCampaignsResponse, len(emailCampaignsResponse), hits.Total, nil
}

func SearchEmailCampaignsByDate(c context.Context, r *http.Request, user models.User) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUserIdQuery := apiSearch.ElasticUserIdQuery{}
	elasticUserIdQuery.Term.UserId = user.Id

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUserIdQuery)

	elasticDateQuery := apiSearch.ElasticSortDataQuery{}
	elasticDateQuery.Date.Order = "desc"
	elasticDateQuery.Date.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticDateQuery)

	return searchEmailCampaigns(c, r, elasticQuery, user)
}
