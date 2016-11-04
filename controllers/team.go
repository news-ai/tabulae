package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/tabulae/models"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getTeam(c context.Context, id int64) (models.Team, error) {
	if id == 0 {
		return models.Team{}, errors.New("datastore: no such entity")
	}

	// Get the team details by id
	var team models.Team
	teamId := datastore.NewKey(c, "Team", "", id, nil)

	err := nds.Get(c, teamId, &team)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Team{}, err
	}

	if !team.Created.IsZero() {
		team.Format(teamId, "teams")
		return team, nil
	}
	return models.Team{}, errors.New("No team by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetTeams(c context.Context, r *http.Request) ([]models.Team, interface{}, int, error) {
	// Now if user is not querying then check
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, 0, err
	}

	if !user.IsAdmin {
		return []models.Team{}, nil, 0, errors.New("Forbidden")
	}

	query := datastore.NewQuery("Team")
	query = constructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, 0, err
	}

	var teams []models.Team
	teams = make([]models.Team, len(ks))
	err = nds.GetMulti(c, ks, teams)
	if err != nil {
		log.Infof(c, "%v", err)
		return teams, nil, 0, err
	}

	for i := 0; i < len(teams); i++ {
		teams[i].Format(ks[i], "teams")
	}

	return teams, nil, len(teams), nil
}
