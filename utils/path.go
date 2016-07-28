package utils

import (
	"encoding/base64"
	"math/rand"
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

// State can be some kind of random generated hash string.
// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
func RandToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
