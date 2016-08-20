package controllers

import (
	"errors"
	"net/http"
	"time"

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

func getAgency(c context.Context, id int64) (models.Agency, error) {
	if id == 0 {
		return models.Agency{}, errors.New("datastore: no such entity")
	}
	// Get the agency by id
	var agency models.Agency
	agencyId := datastore.NewKey(c, "Agency", "", id, nil)
	err := nds.Get(c, agencyId, &agency)
	if err != nil {
		return models.Agency{}, err
	}

	if !agency.Created.IsZero() {
		agency.Id = agencyId.IntID()
		return agency, nil
	}
	return models.Agency{}, errors.New("No agency by this id")
}

/*
* Filter methods
 */

func filterAgency(c context.Context, queryType, query string) (models.Agency, error) {
	ks, err := datastore.NewQuery("Agency").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		return models.Agency{}, err
	}

	if len(ks) == 0 {
		return models.Agency{}, errors.New("No agency by the field " + queryType)
	}

	var agencies []models.Agency
	agencies = make([]models.Agency, len(ks))
	err = nds.GetMulti(c, ks, agencies)
	if err != nil {
		return models.Agency{}, err
	}

	if len(agencies) > 0 {
		agencies[0].Id = ks[0].IntID()
		return agencies[0], nil
	}
	return models.Agency{}, errors.New("No agency by the field " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single agency
func GetAgencies(c context.Context, r *http.Request) ([]models.Agency, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		return []models.Agency{}, err
	}

	if !user.IsAdmin {
		return []models.Agency{}, errors.New("Forbidden")
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	ks, err := datastore.NewQuery("Agency").Limit(limit).Offset(offset).KeysOnly().GetAll(c, nil)
	if err != nil {
		return []models.Agency{}, err
	}

	var agencies []models.Agency
	agencies = make([]models.Agency, len(ks))
	err = nds.GetMulti(c, ks, agencies)
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Agency{}, err
	}

	for i := 0; i < len(agencies); i++ {
		agencies[i].Id = ks[i].IntID()
	}

	return agencies, nil
}

func GetAgency(c context.Context, id string) (models.Agency, error) {
	// Get the details of the current agency
	currentId, err := utils.StringIdToInt(id)
	if err != nil {
		return models.Agency{}, err
	}

	agency, err := getAgency(c, currentId)
	if err != nil {
		return models.Agency{}, err
	}
	return agency, nil
}

/*
* Create methods
 */

func CreateAgencyFromUser(c context.Context, r *http.Request, u *models.User) (models.Agency, error) {
	agencyEmail, err := utils.ExtractAgencyEmail(u.Email)
	if err != nil {
		return models.Agency{}, err
	} else {
		agency, err := FilterAgencyByEmail(c, agencyEmail)
		if err != nil {
			agency = models.Agency{}
			agency.Name, err = utils.ExtractAgencyName(agencyEmail)
			agency.Email = agencyEmail
			agency.Created = time.Now()

			// The person who signs up for the agency at the beginning
			// becomes the defacto administrator until we change.
			agency.Administrators = append(agency.Administrators, u.Id)
			currentUser, err := GetCurrentUser(c, r)
			if err != nil {
				return agency, err
			}
			agency.Create(c, r, currentUser)
		}
		u.Employers = append(u.Employers, agency.Id)
		u.Save(c)
		return agency, nil
	}
	return models.Agency{}, nil
}

/*
* Filter methods
 */

func FilterAgencyByEmail(c context.Context, email string) (models.Agency, error) {
	// Get the id of the current agency
	agency, err := filterAgency(c, "Email", email)
	if err != nil {
		return models.Agency{}, err
	}
	return agency, nil
}
