package utils

import (
	"os"
)

var APIURL = ""

func InitURL() {
	if os.Getenv("RUN_WITH_DEVAPPSERVER") == "1" {
		APIURL = "http://localhost:8080/api"
		return
	}

	if os.Getenv("BASE_URL") != "" {
		APIURL = os.Getenv("BASE_URL")
	} else {
		APIURL = "https://tabulae.newsai.org/api"
	}
}
