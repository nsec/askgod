package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdTimeline(ctx *cli.Context) error {
	// Get the data
	resp := []api.TimelineEntry{}

	err := c.queryStruct("GET", "/timeline", nil, &resp)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	first := true
	for _, entry := range resp {
		if !first {
			fmt.Println("")
		} else {
			first = false
		}

		fmt.Printf("== %d: <%s> %s\n", entry.Team.ID, entry.Team.Country, entry.Team.Name)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Submit time", "Value", "Total"})
		table.SetBorder(false)

		for _, score := range entry.Score {
			table.Append([]string{
				score.SubmitTime.Local().Format(layout),
				fmt.Sprintf("%d", score.Value),
				fmt.Sprintf("%d", score.Total),
			})
		}

		table.Render()
	}

	return nil
}
