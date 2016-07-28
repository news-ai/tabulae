package utils

import (
	"encoding/base64"
	"math/rand"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var APIURL = ""
var salt = os.Getenv("SECRETSALT")

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

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func ValidatePassword(hashedPassword []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}
