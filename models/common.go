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

func ExtractAgencyEmail(email string) (string, error) {
	splitEmail := strings.Split(email, "@")
	if len(splitEmail) > 1 {
		return splitEmail[1], nil
	}
	return "", errors.New("Email is invalid")
}

func ExtractAgencyName(email string) (string, error) {
	splitEmail := strings.Split(email, ".")
	if len(splitEmail) > 1 {
		return strings.Title(splitEmail[0]), nil
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

func StripQueryString(inputUrl string) string {
	u, err := url.Parse(inputUrl)
	if err != nil {
		panic(err)
	}
	u.RawQuery = ""
	return u.String()
}

func UpdateIfNotBlank(initial *string, replace string) {
	if replace != "" {
		*initial = replace
	}
}
