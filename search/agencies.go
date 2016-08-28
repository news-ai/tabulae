package search

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

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

func SearchAgency(c context.Context, r *http.Request, search string) ([]models.Agency, error) {
	search = url.QueryEscape(search)

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	client := urlfetch.Client(c)
	resp, err := client.Get("https://search.newsai.org/agencies/_search?size=" + strconv.Itoa(limit) + "&from=" + strconv.Itoa(offset) + "&q=data.Name:" + search)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var agencyResponse AgencyResponse
	err = decoder.Decode(&agencyResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	agencies := []models.Agency{}
	if agencyResponse.Hits.Total == 0 {
		return agencies, nil
	}

	agencyHits := agencyResponse.Hits.Hits
	for i := 0; i < len(agencyHits); i++ {
		agencyHits[i].Source.Data.Type = "agencies"
		agencies = append(agencies, agencyHits[i].Source.Data)
	}

	return agencies, nil
}
