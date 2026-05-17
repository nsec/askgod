package main

import (
	"cmp"
	"context"
	"os"
	"slices"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (c *client) cmdAdminHistory(ctx context.Context, cmd *cli.Command) error {
	flagIDs := cmd.Int64Slice("flags")

	// Get the scores
	scores := []api.AdminScore{}

	err := c.queryStruct(ctx, "GET", "/scores", nil, &scores)
	if err != nil {
		return err
	}

	slices.SortFunc(scores, func(a api.AdminScore, b api.AdminScore) int {
		return cmp.Compare(a.FlagID, b.FlagID)
	})

	// Get the teams
	teams := []api.AdminTeam{}

	err = c.queryStruct(ctx, "GET", "/teams", nil, &teams)
	if err != nil {
		return err
	}

	// Get the flags
	flags := []api.AdminFlag{}

	err = c.queryStruct(ctx, "GET", "/flags", nil, &flags)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Flag ID", "Flag Description", "Flag Tags", "Team ID", "Team Name", "Team Tags", "Value", "Submit time"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, entry := range scores {
		if !matchesFlagIDs(entry.FlagID, flagIDs) {
			continue
		}

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
