package main

import (
	"os"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// Sorting.
type byFlagID []api.AdminScore

func (a byFlagID) Len() int {
	return len(a)
}

func (a byFlagID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byFlagID) Less(i, j int) bool {
	return a[i].FlagID < a[j].FlagID
}

func (c *client) cmdAdminHistory(_ *cli.Context) error {
	// Get the scores
	scores := []api.AdminScore{}
	err := c.queryStruct("GET", "/scores", nil, &scores)
	if err != nil {
		return err
	}
	sort.Sort(byFlagID(scores))

	// Get the teams
	teams := []api.AdminTeam{}
	err = c.queryStruct("GET", "/teams", nil, &teams)
	if err != nil {
		return err
	}

	// Get the flags
	flags := []api.AdminFlag{}
	err = c.queryStruct("GET", "/flags", nil, &flags)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Flag ID", "Flag Description", "Flag Tags", "Team ID", "Team Name", "Team Tags", "Value", "Submit time"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, entry := range scores {
		// Get the team
		team := api.AdminTeam{}
		for _, t := range teams {
			if t.ID == entry.TeamID {
				team = t

				break
			}
		}

		// Get the team
		flag := api.AdminFlag{}
		for _, t := range flags {
			if t.ID == entry.FlagID {
				flag = t

				break
			}
		}

		teamid := strconv.FormatInt(team.ID, 10)
		if team.Tags["infra"] != "" {
			teamid = team.Tags["infra"]
		}

		table.Append([]string{
			strconv.FormatInt(flag.ID, 10),
			flag.Description,
			utils.PackTags(flag.Tags),
			teamid,
			team.Name,
			utils.PackTags(team.Tags),
			strconv.FormatInt(entry.Value, 10),
			entry.SubmitTime.Local().Format(layout),
		})
	}

	table.Render()

	return nil
}
