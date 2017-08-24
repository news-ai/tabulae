package updates

type EmailSendUpdate struct {
	EmailId int64  `json:"emailid"`
	Method  string `json:"method"`
	SendId  string `json:"sendid"`
}

func init() {

}
