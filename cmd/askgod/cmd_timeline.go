package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdTimeline(_ *cli.Context) error {
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
			_, _ = fmt.Println("")
		} else {
			first = false
		}

		_, _ = fmt.Printf("== %d: <%s> %s\n", entry.Team.ID, entry.Team.Country, entry.Team.Name)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Submit time", "Value", "Total"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		for _, score := range entry.Score {
			table.Append([]string{
				score.SubmitTime.Local().Format(layout),
				strconv.FormatInt(score.Value, 10),
				strconv.FormatInt(score.Total, 10),
			})
		}

		table.Render()
	}

	return nil
}
