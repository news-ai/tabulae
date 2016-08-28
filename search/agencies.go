package search

import (
	"errors"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/models"
)

type AgencyResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int     `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			Type   string  `json:"_type"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Data models.Agency `json:"data"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func SearchAgency(c context.Context, search string) (interface{}, error) {
	client := urlfetch.Client(c)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=" + tkn.AccessToken)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var agencyResponse AgencyResponse
	err = decoder.Decode(&agencyResponse)
	if err != nil {
		return nil, err
	}

	agencies := []models.Agency{}
	if agencyResponse.Hits.Total == 0 {
		return agencies, nil
	}

	agencyHits := agencyResponse.Hits.Hits
	for i := 0; i < len(agencyHits); i++ {
		agencyHits = append(agencyHits, agencyHits[i].Source.Data)
	}

	return agencyHits, nil
}
