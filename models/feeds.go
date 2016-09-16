package models

type Feed struct {
	Base

	FeedURL   string `json:"feedurl"`
	ContactId int64  `json:"contactid"`

	ValidFeed bool `json:"validfeed"`
}

/*
* Private methods
 */

/*
* Create methods
 */

func (f *Feed) Create(c context.Context, r *http.Request, currentUser User) (*Feed, error) {
	f.IsSent = false
	f.CreatedBy = currentUser.Id
	f.Created = time.Now()

	_, err := f.Save(c)
	return f, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (f *Feed) Save(c context.Context) (*Feed, error) {
	// Update the Updated time
	f.Updated = time.Now()

	k, err := nds.Put(c, f.key(c, "Feed"), f)
	if err != nil {
		return nil, err
	}
	f.Id = k.IntID()
	return f, nil
}
