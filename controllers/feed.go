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

	"github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/sync"

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

func filterFeeds(c context.Context, r *http.Request, queryType, query string) ([]models.Feed, error) {
	// Get an contact by a query type
	ks, err := datastore.NewQuery("File").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, err
	}

	var feeds []models.Feed
	feeds = make([]models.Feed, len(ks))
	err = nds.GetMulti(c, ks, feeds)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, err
	}

	if len(feeds) > 0 {
		for i := 0; i < len(feeds); i++ {
			feeds[i].Format(ks[i], "feeds")
		}

		return feeds, nil

	}
	return []models.Feed{}, errors.New("No feed by this " + queryType)
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

func GetFeedById(c context.Context, r *http.Request, id int64) (models.Feed, interface{}, error) {
	feed, err := getFeed(c, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}
	return feed, nil, nil
}

func GetFeeds(c context.Context, r *http.Request) ([]models.Feed, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, nil, 0, 0, err
	}

	query := datastore.NewQuery("Feed").Filter("CreatedBy =", user.Id)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, nil, 0, 0, err
	}

	var feeds []models.Feed
	feeds = make([]models.Feed, len(ks))

	err = nds.GetMulti(c, ks, feeds)
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Feed{}, nil, 0, 0, err
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].Format(ks[i], "feeds")
	}

	return feeds, nil, len(feeds), 0, nil
}

func GetFeedsByResourceId(c context.Context, r *http.Request, resouceName string, resourceId int64) ([]models.Feed, error) {
	query := datastore.NewQuery("Feed").Filter(resouceName+" =", resourceId)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Feed{}, err
	}

	var feeds []models.Feed
	feeds = make([]models.Feed, len(ks))

	err = nds.GetMulti(c, ks, feeds)
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Feed{}, err
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].Format(ks[i], "feeds")
	}

	return feeds, nil
}

func FilterFeeds(c context.Context, r *http.Request, queryType, query string) ([]models.Feed, error) {
	// User has to be logged in
	_, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		return []models.Feed{}, err
	}

	return filterFeeds(c, r, queryType, query)
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

	currentUser, err := controllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return feed, nil, err
	}

	// Check if the feed already exists for a particular contact by the same user
	query := datastore.NewQuery("Feed").Filter("FeedURL =", feed.FeedURL).Filter("CreatedBy =", currentUser.Id).Filter("ContactId =", feed.ContactId)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	if len(ks) > 0 {
		return models.Feed{}, nil, errors.New("Feed already exits for the contact")
	}

	baseDomain, err := utilities.NormalizeUrl(feed.FeedURL)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	publicationName, err := utilities.GetTitleFromHTTPRequest(c, baseDomain)
	if err != nil {
		publicationName, err = utilities.GetDomainName(baseDomain)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Feed{}, nil, err
		}
	}

	publication, err := FindOrCreatePublication(c, r, publicationName, baseDomain)
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

	// Run new feed through pub/sub
	sync.NewRSSFeedSync(r, feed.FeedURL, feed.PublicationId)

	return feed, nil, nil
}

/*
* Delete methods
 */

func DeleteFeed(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	keyID := datastore.NewKey(c, "Feed", "", currentId, nil)
	err = nds.Delete(c, keyID)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Feed{}, nil, err
	}

	return nil, nil, nil
}
