package controllers

import (
	"errors"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

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
