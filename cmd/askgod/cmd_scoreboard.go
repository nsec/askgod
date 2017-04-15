package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdScoreboard(ctx *cli.Context) error {
	// Get the data
	resp := []api.ScoreboardEntry{}

	err := c.queryStruct("GET", "/scoreboard", nil, &resp)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Team", "Points", "Last submit"})
	table.SetBorder(false)

	for _, entry := range resp {
		table.Append([]string{
			fmt.Sprintf("<%s> %s ", entry.Team.Country, entry.Team.Name),
			fmt.Sprintf("%d", entry.Value),
			entry.LastSubmitTime.Local().Format(layout),
		})
	}

	table.Render()

	return nil
}
