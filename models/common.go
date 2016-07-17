package models

import (
	"errors"
	"strconv"
	"strings"
)

func StringIdToInt(id string) (int64, error) {
	currentId, err := strconv.ParseInt(id, 10, 32)
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
