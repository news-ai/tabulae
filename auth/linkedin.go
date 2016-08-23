package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

var (
	linkedinOauthConfig = &oauth2.Config{
		RedirectURL:  "http://tabulae.newsai.org/api/auth/linkedincallback",
		ClientID:     os.Getenv("LINKEDINAUTHKEY"),
		ClientSecret: os.Getenv("LINKEDINAUTHSECRET"),
		Endpoint:     linkedin.Endpoint,
	}
)