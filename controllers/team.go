package controllers

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
		team.Format(publicationId, "teams")
		return team, nil
	}
	return models.Team{}, errors.New("No team by this id")
}
