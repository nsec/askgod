package main

import (
	"context"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdHistory(ctx context.Context, cmd *cli.Command) error {
	// Get the data
	resp := []api.Flag{}

	if cmd.NArg() > 0 {
		flag := api.Flag{}

		err := c.queryStruct(ctx, "GET", "/team/flags/"+cmd.Args().Get(0), nil, &flag)
		if err != nil {
			return err
		}

		if cmd.NArg() > 1 {
			for _, arg := range cmd.Args().Slice()[1:] {
				err := setStructKey(&flag, arg)
				if err != nil {
					return err
				}
			}

			err = c.queryStruct(ctx, "PUT", "/team/flags/"+cmd.Args().Get(0), flag.FlagPut, nil)
			if err != nil {
				return err
			}

			return nil
		}

		resp = append(resp, flag)
	} else {
		err := c.queryStruct(ctx, "GET", "/team/flags", nil, &resp)
		if err != nil {
			return err
		}
	}

	const layout = "2006/01/02 15:04"

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Description", "Value", "Timestamp", "Source", "Message", "Notes"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, flag := range resp {
		table.Append([]string{
			strconv.FormatInt(flag.ID, 10),
			flag.Description,
			strconv.FormatInt(flag.Value, 10),
			flag.SubmitTime.Local().Format(layout),
			flag.Source,
			flag.ReturnString,
			flag.Notes,
		})
	}

	table.Render()

	return nil
}
