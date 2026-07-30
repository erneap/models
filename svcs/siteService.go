package svcs

import (
	"sort"
	"strings"

	"github.com/erneap/models/employees"
	"github.com/erneap/models/sites"
)

// Every service will have functions for completing the CRUD functions
// the retrieve functions will be for individual site and the whole list of
// tea's sotes.

func CreateSite(teamid string, id, name string) (*sites.Site, error) {
	team, err := GetTeam(teamid)
	if err != nil {
		return nil, err
	}

	var answer *sites.Site

	for _, site := range team.Sites {
		if strings.EqualFold(site.ID, id) || strings.EqualFold(site.Name, name) {
			answer = &site
		}
	}
	if answer == nil {
		answer = &sites.Site{
			ID:   id,
			Name: name,
		}
		team.Sites = append(team.Sites, *answer)
	} else {
		answer.ID = id
		answer.Name = name
	}
	UpdateTeam(team)

	return answer, nil
}

func GetSite(teamid, siteid string) (*sites.Site, error) {
	team, err := GetTeam(teamid)
	if err != nil {
		return nil, err
	}

	var answer sites.Site
	for _, site := range team.Sites {
		if strings.EqualFold(site.ID, siteid) {
			answer = site
			emps, _ := GetEmployees(teamid, siteid)
			answer.Employees = append(site.Employees, emps...)
			sort.Sort(employees.ByEmployees(answer.Employees))
		}
	}
	return &answer, nil
}

func GetSites(teamid string) ([]sites.Site, error) {
	team, err := GetTeam(teamid)
	if err != nil {
		return nil, err
	}

	sort.Sort(sites.BySites(team.Sites))
	return team.Sites, nil
}

func UpdateSite(teamid string, nsite sites.Site) error {
	team, err := GetTeam(teamid)
	if err != nil {
		return err
	}

	for s, site := range team.Sites {
		if strings.EqualFold(site.ID, nsite.ID) {
			team.Sites[s] = nsite
		}
	}
	return UpdateTeam(team)
}

func DeleteSite(teamid, siteid string) error {
	team, err := GetTeam(teamid)
	if err != nil {
		return err
	}

	found := -1
	for s := 0; s < len(team.Sites) && found < 0; s++ {
		if strings.EqualFold(team.Sites[s].ID, siteid) {
			found = s
		}
	}
	if found >= 0 {
		team.Sites = append(team.Sites[:found], team.Sites[found+1:]...)
	}
	return UpdateTeam(team)
}
