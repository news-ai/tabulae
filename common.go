package tabulae

type User struct {
	Id       int64  `json:"id" datastore:"-"`
	GoogleId string `json:"googleid"`

	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	Agency int64 `json:"agencyid" datastore:"-"`

	Created time.Time `json:"created"`
}

type Agency struct {
	Id int64 `json:"id" datastore:"-"`

	AgencyName string `json:"agencyname"`
}
