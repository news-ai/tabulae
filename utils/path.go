package utils

import (
	"os"
)

var APIURL = ""

func InitURL() string {
	if os.Getenv("RUN_WITH_DEVAPPSERVER") == "1" {
		APIURL = "http://localhost:8080/api"
	} else {
		APIURL = "http://tabulae.newsai.org/api"
	}

	return APIURL
}
