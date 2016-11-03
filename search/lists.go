package search

import (
	"time"
)

type Lists struct {
	Type string `json:"type"`

	Archived   bool      `json:"archived"`
	FileUpload int64     `json:"fileupload"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	CreatedBy  int64     `json:"createdby"`
	Client     string    `json:"client"`
	Id         int64     `json:"id"`
}

func (l *Lists) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := models.SetField(l, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
