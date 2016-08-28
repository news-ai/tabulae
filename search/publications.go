package search

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/news-ai/tabulae/models"
)

type PublicationResponse struct {
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
				Data models.Publication `json:"data"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func SearchPublication(c context.Context, r *http.Request, search string) ([]models.Publication, error) {
	search = url.QueryEscape(search)
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	client := urlfetch.Client(c)
	resp, err := client.Get("https://search.newsai.org/publications/_search?size=" + strconv.Itoa(limit) + "&from=" + strconv.Itoa(offset) + "&q=data.Name:" + search)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var publicationResponse PublicationResponse
	err = decoder.Decode(&publicationResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	publications := []models.Publication{}
	if publicationResponse.Hits.Total == 0 {
		return publications, nil
	}

	publicationHits := publicationResponse.Hits.Hits
	for i := 0; i < len(publicationHits); i++ {
		publicationHits[i].Source.Data.Type = "publications"
		publications = append(publications, publicationHits[i].Source.Data)
	}

	return publications, nil
}
