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

func getTemplate(c context.Context, id int64) (models.Template, error) {
	if id == 0 {
		return models.Template{}, errors.New("datastore: no such entity")
	}
	// Get the publication details by id
	var template models.Template
	templateId := datastore.NewKey(c, "Template", "", id, nil)

	err := nds.Get(c, templateId, &template)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, err
	}

	if !template.Created.IsZero() {
		template.Format(templateId, "templates")
		return template, nil
	}
	return models.Template{}, errors.New("No template by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetTemplate(c context.Context, r *http.Request, id string) (models.Template, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}

	template, err := getTemplate(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}

	return template, nil, nil
}

func GetTemplates(c context.Context, r *http.Request) ([]models.Template, interface{}, int, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Template{}, nil, 0, err
	}

	query := datastore.NewQuery("Template").Filter("CreatedBy =", user.Id)
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Template{}, nil, 0, err
	}

	var templates []models.Template
	templates = make([]models.Template, len(ks))

	err = nds.GetMulti(c, ks, templates)
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Template{}, nil, 0, err
	}

	for i := 0; i < len(templates); i++ {
		templates[i].Format(ks[i], "templates")
	}

	return templates, nil, len(templates), nil
}

/*
* Create methods
 */

func CreateTemplate(c context.Context, r *http.Request) (models.Template, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var template models.Template
	err := decoder.Decode(buf, &template)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return template, nil, err
	}

	// Create template
	_, err = template.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}

	return template, nil, nil
}

/*
* Update methods
 */

func UpdateTemplate(c context.Context, r *http.Request, id string) (models.Template, interface{}, error) {
	// Get the details of the current template
	template, _, err := GetTemplate(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}
	if template.CreatedBy != user.Id {
		return models.Template{}, nil, errors.New("Forbidden")
	}

	decoder := ffjson.NewDecoder()
	buf, _ := ioutil.ReadAll(r.Body)
	var updatedTemplate models.Template
	err = decoder.Decode(buf, &updatedTemplate)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Template{}, nil, err
	}

	utilities.UpdateIfNotBlank(&template.Name, updatedTemplate.Name)
	utilities.UpdateIfNotBlank(&template.Subject, updatedTemplate.Subject)
	utilities.UpdateIfNotBlank(&template.Body, updatedTemplate.Body)

	template.Save(c)
	return template, nil, nil
}
