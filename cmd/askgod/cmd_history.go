package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdHistory(ctx *cli.Context) error {
	// Get the data
	resp := []api.Flag{}

	err := c.queryStruct("GET", "/team/flags", nil, &resp)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Flag", "Value", "Timestamp", "Message", "Notes"})
	table.SetBorder(false)

	for _, flag := range resp {
		table.Append([]string{
			fmt.Sprintf("%d", flag.ID),
			flag.Flag,
			fmt.Sprintf("%d", flag.Value),
			flag.SubmitTime.Local().Format(layout),
			flag.ReturnString,
			flag.Notes,
		})
	}

	table.Render()

	return nil
}
