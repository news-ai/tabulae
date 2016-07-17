package models

import (
	"strconv"
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
