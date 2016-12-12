package enrichment

import (
	"net/http"
	"os"
)

func GetProfileFromEmail(r *http.Request, c context.Context, email string) {
	URL := "https://api.fullcontact.com/v2/person.json?email=" + email
	req, _ := http.NewRequest("GET", URL, nil)

	req.Header.Add("X-FullContact-APIKey", os.Getenv("FULLCONTACT_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return "", "", err
	}
}
