package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

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

func GetTeam(c context.Context, id string) (models.Team, interface{}, error) {
	// Get the details of the current team
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Team{}, nil, err
	}

	team, err := getTeam(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Team{}, nil, err
	}
	return team, nil, nil
}

/*
* Create methods
 */

func CreateTeam(c context.Context, r *http.Request) ([]models.Team, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	var team models.Team
	err = decoder.Decode(buf, &team)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var teams []models.Team

		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &teams)

		if err != nil {
			log.Errorf(c, "%v", err)
			return []models.Team{}, nil, err
		}

		newTeams := []models.Team{}
		for i := 0; i < len(teams); i++ {
			_, err = teams[i].Create(c, r, currentUser)
			if err != nil {
				log.Errorf(c, "%v", err)
				return []models.Team{}, nil, err
			}
			newTeams = append(newTeams, teams[i])
		}

		return newTeams, nil, err
	}

	// Create team
	_, err = team.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, err
	}
	return []models.Team{team}, nil, nil
}
