package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getPublication(c context.Context, id int64) (models.Publication, error) {
	if id == 0 {
		return models.Publication{}, errors.New("datastore: no such entity")
	}
	// Get the publication details by id
	var publication models.Publication
	publicationId := datastore.NewKey(c, "Publication", "", id, nil)

	err := nds.Get(c, publicationId, &publication)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, err
	}

	if !publication.Created.IsZero() {
		publication.Format(publicationId, "publications")
		return publication, nil
	}
	return models.Publication{}, errors.New("No publication by this id")
}

/*
* Filter methods
 */

func filterPublication(c context.Context, queryType, query string) (models.Publication, error) {
	// Get a publication by the URL
	ks, err := datastore.NewQuery("Publication").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, err
	}

	var publications []models.Publication
	publications = make([]models.Publication, len(ks))
	err = nds.GetMulti(c, ks, publications)
	if err != nil {
		log.Infof(c, "%v", err)
		return models.Publication{}, err
	}

	if len(publications) > 0 {
		publications[0].Format(ks[0], "publications")
		return publications[0], nil
	}
	return models.Publication{}, errors.New("No publication by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetPublications(c context.Context, r *http.Request) ([]models.Publication, interface{}, int, error) {
	// If user is querying then it is not denied by the server
	queryField := gcontext.Get(r, "q").(string)
	if queryField != "" {
		publications, err := search.SearchPublication(c, r, queryField)
		if err != nil {
			return []models.Publication{}, nil, 0, err
		}
		return publications, nil, len(publications), nil
	}

	// Now if user is not querying then check
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Publication{}, nil, 0, err
	}

	if !user.IsAdmin {
		return []models.Publication{}, nil, 0, errors.New("Forbidden")
	}

	query := datastore.NewQuery("Publication")
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Publication{}, nil, 0, err
	}

	var publications []models.Publication
	publications = make([]models.Publication, len(ks))
	err = nds.GetMulti(c, ks, publications)
	if err != nil {
		log.Infof(c, "%v", err)
		return publications, nil, 0, err
	}

	for i := 0; i < len(publications); i++ {
		publications[i].Format(ks[i], "publications")
	}
	return publications, nil, len(publications), nil
}

func GetPublication(c context.Context, id string) (models.Publication, interface{}, error) {
	// Get a publication by id
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, err
	}

	publication, err := getPublication(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, err
	}
	return publication, nil, nil
}

func GetHeadlinesForPublication(c context.Context, r *http.Request, id string) (interface{}, interface{}, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	headlines, err := search.SearchHeadlinesByPublicationId(c, r, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, 0, err
	}

	return headlines, nil, len(headlines), nil
}

func GetEnrichCompanyProfile(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	publication, err := getPublication(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	if publication.Url == "" {
		return nil, nil, errors.New("Publication has no URL")
	}

	publicationUrl, err := url.Parse(publication.Url)
	if err != nil {
		return nil, nil, err
	}

	publicationDetail, err := search.SearchCompanyDatabase(c, r, publicationUrl.Host)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, nil, err
	}

	return publicationDetail, nil, nil
}

/*
* Create methods
 */

func CreatePublication(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, interface{}, int, error) {
	// Parse JSON
	buf, _ := ioutil.ReadAll(r.Body)

	decoder := ffjson.NewDecoder()
	var publication models.Publication
	err := decoder.Decode(buf, &publication)

	if err != nil {
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Publication{}, nil, 0, err
		}

		var publications []models.Publication
		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &publications)

		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Publication{}, nil, 0, err
		}

		newPublications := []models.Publication{}
		for i := 0; i < len(publications); i++ {
			_, err = publications[i].Validate(c)
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Publication{}, nil, 0, err
			}

			presentPublication, _, err := FilterPublicationByName(c, publications[i].Name)
			if err != nil {
				_, err = publications[i].Create(c, r, currentUser)
				if err != nil {
					log.Errorf(c, "%v", err)
					return []models.Publication{}, nil, 0, err
				}
				sync.ResourceSync(r, publications[i].Id, "Publication", "create")
				newPublications = append(newPublications, publications[i])
			} else {
				newPublications = append(newPublications, presentPublication)
			}
		}
		return newPublications, nil, len(newPublications), err
	}

	_, err = publication.Validate(c)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, 0, err
	}

	presentPublication, _, err := FilterPublicationByName(c, publication.Name)
	if err != nil {
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Publication{}, nil, 0, err
		}
		// Create publication
		_, err = publication.Create(c, r, currentUser)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Publication{}, nil, 0, err
		}
		sync.ResourceSync(r, publication.Id, "Publication", "create")
		return publication, nil, 1, nil
	}
	return presentPublication, nil, 1, nil
}

/*
* Update methods
 */

func UpdatePublication(c context.Context, r *http.Request, id string) (models.Publication, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, err
	}

	publication, err := getPublication(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	buf, _ := ioutil.ReadAll(r.Body)
	var updatedPublication models.Publication
	err = decoder.Decode(buf, &updatedPublication)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, err
	}

	if publication.Url == "" && updatedPublication.Url != "" {
		publication.Url = updatedPublication.Url
		publication.Save(c)
	}

	return publication, nil, nil
}

func FindOrCreatePublication(c context.Context, r *http.Request, name string) (models.Publication, error) {
	name = strings.Trim(name, " ")
	publication, _, err := FilterPublicationByName(c, name)
	if err != nil {
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Publication{}, err
		}

		var newPublication models.Publication
		newPublication.Name = name
		_, err = newPublication.Create(c, r, currentUser)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.Publication{}, err
		}

		sync.ResourceSync(r, newPublication.Id, "Publication", "create")
		return newPublication, nil
	}

	return publication, nil
}

/*
* Filter methods
 */

func FilterPublicationByUrl(c context.Context, url string) (models.Publication, error) {
	// Get the id of the current publication
	publication, err := filterPublication(c, "Url", url)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, err
	}
	return publication, nil
}

func FilterPublicationByName(c context.Context, name string) (models.Publication, interface{}, error) {
	// Get the id of the current publication
	publication, err := filterPublication(c, "Name", name)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Publication{}, nil, err
	}
	return publication, nil, nil
}
