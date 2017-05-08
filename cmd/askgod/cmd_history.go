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

	if ctx.NArg() > 0 {
		flag := api.Flag{}
		err := c.queryStruct("GET", fmt.Sprintf("/team/flags/%s", ctx.Args().Get(0)), nil, &flag)
		if err != nil {
			return err
		}

		if ctx.NArg() > 1 {
			for _, arg := range ctx.Args()[1:] {
				err := setStructKey(&flag, arg)
				if err != nil {
					return err
				}
			}

			err = c.queryStruct("PUT", fmt.Sprintf("/team/flags/%s", ctx.Args().Get(0)), flag.FlagPut, nil)
			if err != nil {
				return err
			}

			return nil
		}

		resp = append(resp, flag)
	} else {
		err := c.queryStruct("GET", "/team/flags", nil, &resp)
		if err != nil {
			return err
		}
	}

	const layout = "2006/01/02 15:04"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Description", "Value", "Timestamp", "Message", "Notes"})
	table.SetBorder(false)

	for _, flag := range resp {
		table.Append([]string{
			fmt.Sprintf("%d", flag.ID),
			flag.Description,
			fmt.Sprintf("%d", flag.Value),
			flag.SubmitTime.Local().Format(layout),
			flag.ReturnString,
			flag.Notes,
		})
	}

	table.Render()

	return nil
}
