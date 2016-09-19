package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"

	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getFeed(c context.Context, id int64) (models.Feed, error) {
	if id == 0 {
		return models.Feed{}, errors.New("datastore: no such entity")
	}
	// Get the publication details by id
	var feed models.Feed
	feedId := datastore.NewKey(c, "Feed", "", id, nil)

	err := nds.Get(c, feedId, &feed)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, err
	}

	if !feed.Created.IsZero() {
		feed.Format(feedId, "feeds")
		return feed, nil
	}
	return models.Feed{}, errors.New("No feed by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetFeed(c context.Context, r *http.Request, id string) (models.Feed, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	feed, err := getFeed(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	return feed, nil, nil
}

func GetFeeds(c context.Context, r *http.Request) ([]models.Feed, interface{}, int, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, nil, 0, err
	}

	query := datastore.NewQuery("Feed").Filter("CreatedBy =", user.Id)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, nil, 0, err
	}

	var feeds []models.Feed
	feeds = make([]models.Feed, len(ks))

	err = nds.GetMulti(c, ks, feeds)
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Feed{}, nil, 0, err
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].Format(ks[i], "feeds")
	}

	return feeds, nil, len(feeds), nil
}

/*
* Create methods
 */

func CreateFeed(c context.Context, r *http.Request) (models.Feed, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var feed models.Feed
	err := decoder.Decode(buf, &feed)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return feed, nil, err
	}

	baseDomain, err := utilities.NormalizeUrl(feed.FeedURL)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	publicationName, err := utilities.GetTitleFromHTTPRequest(c, baseDomain)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	publication, err := FindOrCreatePublication(c, r, publicationName)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	publication.Url = baseDomain
	publication.Save(c)

	feed.PublicationId = publication.Id

	// Create feed
	_, err = feed.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	return feed, nil, nil
}
