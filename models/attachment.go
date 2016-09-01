package models

type Attachment struct {
	Base

	FileName string `json:"filename"`
	EmailId  int64  `json:"emailid"`

	FileExists bool `json:"fileexists"`
}
