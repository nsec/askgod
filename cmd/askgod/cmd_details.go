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

func (c *client) cmdDetails(ctx *cli.Context) error {
	// Get the data
	resp := api.Team{}

	err := c.queryStruct("GET", "/team", nil, &resp)
	if err != nil {
		return err
	}

	// Process any field update
	if ctx.NArg() > 0 {
		v := reflect.ValueOf(&resp)

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

			field.SetString(fields[1])
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
