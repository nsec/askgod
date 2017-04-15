package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdAdminAddScore(ctx *cli.Context) error {
	score := api.AdminScorePost{}

	if ctx.NArg() > 0 {
		v := reflect.ValueOf(&score)

		for _, arg := range ctx.Args() {
			fields := strings.SplitN(arg, "=", 2)
			if len(fields) != 2 {
				return fmt.Errorf("Bad key=value input: %s", arg)
			}

			field := v.Elem().FieldByNameFunc(func(name string) bool {
				if strings.ToLower(name) == strings.ToLower(fields[0]) {
					return true
				}

				return false
			})

			if !field.IsValid() {
				return fmt.Errorf("Invalid key: %s", fields[0])
			}

			if field.Type() == reflect.TypeOf("") {
				field.SetString(fields[1])
			} else if field.Type() == reflect.TypeOf(int64(0)) {
				intValue, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return err
				}

				field.SetInt(intValue)
			} else {
				return fmt.Errorf("Unsupported type for key: %s", fields[0])
			}
		}
	}

	err := c.queryStruct("POST", "/scores", score, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminDeleteScore(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	err := c.queryStruct("DELETE", fmt.Sprintf("/scores/%s", ctx.Args().Get(0)), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminListScores(ctx *cli.Context) error {
	// Get the data
	resp := []api.AdminScore{}

	err := c.queryStruct("GET", "/scores", nil, &resp)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "TeamID", "FlagID", "Value", "Submit time", "Notes"})
	table.SetBorder(false)

	for _, entry := range resp {
		table.Append([]string{
			fmt.Sprintf("%d", entry.ID),
			fmt.Sprintf("%d", entry.TeamID),
			fmt.Sprintf("%d", entry.FlagID),
			fmt.Sprintf("%d", entry.Value),
			entry.SubmitTime.Local().Format(layout),
			entry.Notes,
		})
	}

	table.Render()

	return nil
}

func (c *client) cmdAdminUpdateScore(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	score := api.AdminScore{}
	err := c.queryStruct("GET", fmt.Sprintf("/scores/%s", ctx.Args().Get(0)), nil, &score)
	if err != nil {
		return err
	}

	if ctx.NArg() > 1 {
		v := reflect.ValueOf(&score)

		for _, arg := range ctx.Args()[1:] {
			fields := strings.SplitN(arg, "=", 2)
			if len(fields) != 2 {
				return fmt.Errorf("Bad key=value input: %s", arg)
			}

			field := v.Elem().FieldByNameFunc(func(name string) bool {
				if strings.ToLower(name) == strings.ToLower(fields[0]) {
					return true
				}

				return false
			})

			if !field.IsValid() {
				return fmt.Errorf("Invalid key: %s", fields[0])
			}

			if field.Type() == reflect.TypeOf("") {
				field.SetString(fields[1])
			} else if field.Type() == reflect.TypeOf(int64(0)) {
				intValue, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return err
				}

				field.SetInt(intValue)
			} else {
				return fmt.Errorf("Unsupported type for key: %s", fields[0])
			}
		}
	}

	err = c.queryStruct("PUT", fmt.Sprintf("/scores/%s", ctx.Args().Get(0)), score.AdminScorePut, nil)
	if err != nil {
		return err
	}

	return nil
}
