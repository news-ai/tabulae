package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
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
		return models.Template{}, err
	}

	if !template.Created.IsZero() {
		template.Id = templateId.IntID()
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

func GetTemplate(c context.Context, r *http.Request, id string) (models.Template, error) {
	// Get the details of the current user
	currentId, err := utils.StringIdToInt(id)
	if err != nil {
		return models.Template{}, err
	}

	template, err := getTemplate(c, currentId)
	if err != nil {
		return models.Template{}, err
	}
	return template, nil
}

func GetTemplates(c context.Context, r *http.Request) ([]models.Template, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Template{}, err
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	ks, err := datastore.NewQuery("Template").Filter("CreatedBy =", user.Id).Limit(limit).Offset(offset).KeysOnly().GetAll(c, nil)
	if err != nil {
		return []models.Template{}, err
	}

	var templates []models.Template
	templates = make([]models.Template, len(ks))

	err = nds.GetMulti(c, ks, templates)
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Template{}, err
	}

	for i := 0; i < len(templates); i++ {
		templates[i].Id = ks[i].IntID()
	}

	return templates, nil
}

/*
* Create methods
 */

func CreateTemplate(c context.Context, r *http.Request) (models.Template, error) {
	decoder := json.NewDecoder(r.Body)
	var template models.Template
	err := decoder.Decode(&template)
	if err != nil {
		return models.Template{}, err
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		return template, err
	}

	// Create template
	_, err = template.Create(c, r, currentUser)
	if err != nil {
		return models.Template{}, err
	}

	return template, nil
}

/*
* Update methods
 */

func UpdateTemplate(c context.Context, r *http.Request, id string) (models.Template, error) {
	// Get the details of the current template
	template, err := GetTemplate(c, r, id)
	if err != nil {
		return models.Template{}, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return models.Template{}, err
	}
	if template.CreatedBy != user.Id {
		return models.Template{}, errors.New("Forbidden")
	}

	decoder := json.NewDecoder(r.Body)
	var updatedTemplate models.Template
	err = decoder.Decode(&updatedTemplate)
	if err != nil {
		return models.Template{}, err
	}

	utils.UpdateIfNotBlank(&template.Subject, updatedTemplate.Subject)
	utils.UpdateIfNotBlank(&template.Body, updatedTemplate.Body)

	template.Save(c)
	return template, nil
}
