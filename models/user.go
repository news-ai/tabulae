package models

import (
	"github.com/news-ai/tabulae"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func defaultUserList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "UserList", "default", 0, nil)
}

func (u *tabulae.User) key(c appengine.Context) *datastore.Key {
	if u.Id == 0 {
		u.Created = time.Now()
		return datastore.NewIncompleteKey(c, "User", defaultUserList(c))
	}
	return datastore.NewKey(c, "User", "", u.Id, defaultUserList(c))
}

func (u *tabulae.User) save(c appengine.Context) (*tabulae.User, error) {
	k, err := datastore.Put(c, u.key(c), u)
	if err != nil {
		return nil, err
	}
	u.Id = k.IntID()
	return u, nil
}

func decodeUser(r io.ReadCloser) (*tabulae.User, error) {
	defer r.Close()
	var user tabulae.User
	err := json.NewDecoder(r).Decode(&user)
	return &user, err
}

func getCurrentUser(c appengine.Context) (tabulae.User, error) {
	user := []tabulae.User{}
	ks, err := datastore.NewQuery("User").Ancestor(defaultUserList(c)).Order("Created").GetAll(c, &user)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(todos); i++ {
		user[i].Id = ks[i].IntID()
	}
	return user[0], nil
}
