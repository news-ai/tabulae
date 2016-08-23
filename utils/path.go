package utils

import (
	"os"
)

var APIURL = ""

func InitURL() {
	if os.Getenv("BASE_URL") != "" {
		APIURL = os.Getenv("BASE_URL")
		return
	}

	if os.Getenv("RUN_WITH_DEVAPPSERVER") == "1" {
		APIURL = "http://localhost:8080/api"
	} else {
		APIURL = "https://tabulae.newsai.org/api"
	}
}
