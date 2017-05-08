package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdDetails(ctx *cli.Context) error {
	// Get the data
	resp := api.Team{}

	err := c.queryStruct("GET", "/team", nil, &resp)
	if err != nil {
		return err
	}

	// Process any field update
	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args() {
			err := setStructKey(&resp, arg)
			if err != nil {
				return err
			}
		}

		// Update the team
		err = c.queryStruct("PUT", "/team", resp.TeamPut, nil)
		if err != nil {
			return err
		}

		return nil
	}

	// Render the result
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	val := func(in string) string {
		if in == "" {
			return "<unset>"
		}

		return in
	}

	table.Append([]string{"NAME", val(resp.Name)})
	table.Append([]string{"COUNTRY", val(resp.Country)})
	table.Append([]string{"WEBSITE", val(resp.Website)})

	table.Render()

	return nil
}
