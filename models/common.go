package models

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

func StringIdToInt(id string) (int64, error) {
	currentId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return currentId, nil
}

func IntIdToString(id int64) string {
	currentId := strconv.FormatInt(id, 10)
	return currentId
}

func GetAgencyEmail(email string) (string, error) {
	splitEmail := strings.Split(email, "@")
	if len(splitEmail) > 1 {
		return splitEmail[1], nil
	}
	return "", errors.New("Email is invalid")
}

func GetAgencyName(email string) (string, error) {
	splitEmail := strings.Split(email, ".")
	if len(splitEmail) > 1 {
		return splitEmail[0], nil
	}
	return "", errors.New("Name is invalid")
}

func NormalizeUrl(initialUrl string) (string, error) {
	u, err := url.Parse(initialUrl)
	if err != nil {
		return "", err
	}
	urlHost := strings.Replace(u.Host, "www.", "", 1)
	return u.Scheme + "://" + urlHost, nil
}
