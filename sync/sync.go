package sync

import (
	"bytes"
	"encoding/json"
	"net/http"

	"appengine"
	"appengine/urlfetch"
)

type LinkedInData struct {
	Current []struct {
		Date     string `json:"date"`
		Position string `json:"position"`
		Employer string `json:"employer"`
	} `json:"current"`
	Past []struct {
		Date     string `json:"date"`
		Position string `json:"position"`
		Employer string `json:"employer"`
	} `json:"past"`
}

func LinkedInSync(r *http.Request, contactId int64, contactLinkedIn string) LinkedInData {
	c := appengine.NewContext(r)
	url := "http://influencer.newsai.org/"

	var jsonStr = []byte(`{"url": "` + contactLinkedIn + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	client := urlfetch.Client(c)
	resp, err := client.Do(req)

	if err != nil {
		return LinkedInData{}
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var linkedInData LinkedInData
	err = decoder.Decode(&linkedInData)
	if err != nil {
		return LinkedInData{}
	}

	return linkedInData
}
