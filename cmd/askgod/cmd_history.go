package main

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdHistory(ctx *cli.Context) error {
	// Get the data
	resp := []api.Flag{}

	if ctx.NArg() > 0 {
		flag := api.Flag{}
		err := c.queryStruct("GET", "/team/flags/"+ctx.Args().Get(0), nil, &flag)
		if err != nil {
			return err
		}

		if ctx.NArg() > 1 {
			for _, arg := range ctx.Args().Slice()[1:] {
				err := setStructKey(&flag, arg)
				if err != nil {
					return err
				}
			}

			err = c.queryStruct("PUT", "/team/flags/"+ctx.Args().Get(0), flag.FlagPut, nil)
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
	table.SetAutoWrapText(false)

	for _, flag := range resp {
		table.Append([]string{
			strconv.FormatInt(flag.ID, 10),
			flag.Description,
			strconv.FormatInt(flag.Value, 10),
			flag.SubmitTime.Local().Format(layout),
			flag.ReturnString,
			flag.Notes,
		})
	}

	table.Render()

	return nil
}
