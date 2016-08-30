package models

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/utils"
)

type Publication struct {
	Base

	Name string `json:"name"`
	Url  string `json:"url"`

	// Social information
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	MuckRack  string `json:"-"`
	Website   string `json:"website"`
	Blog      string `json:"blog"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new publication into App Engine
func (p *Publication) Create(c context.Context, r *http.Request, currentUser User) (*Publication, error) {
	p.CreatedBy = currentUser.Id
	p.Created = time.Now()

	_, err := p.Save(c)
	return p, err
}

/*
* Update methods
 */

// Function to save a new publication into App Engine
func (p *Publication) Save(c context.Context) (*Publication, error) {
	// Update the Updated time
	p.Updated = time.Now()

	// Save the object
	k, err := nds.Put(c, p.key(c, "Publication"), p)
	if err != nil {
		return nil, err
	}
	p.Id = k.IntID()
	return p, nil
}

func (p *Publication) Validate(c context.Context) (*Publication, error) {
	// Validate Fields
	if p.Name == "" {
		return p, errors.New("Missing fields")
	}

	// Format URL properly
	if p.Url != "" {
		normalizedUrl, err := utils.NormalizeUrl(p.Url)
		if err != nil {
			return p, err
		}
		p.Url = normalizedUrl
	}
	return p, nil
}

func (p *Publication) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := SetField(p, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
