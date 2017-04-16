package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (c *client) cmdAdminAddTeam(ctx *cli.Context) error {
	team := api.AdminTeamPost{}

	if ctx.NArg() > 0 {
		v := reflect.ValueOf(&team)

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
			} else if field.Type() == reflect.TypeOf(map[string]string{}) {
				tags, err := utils.ParseTags(fields[1])
				if err != nil {
					return err
				}

				tagsValue := reflect.ValueOf(tags)
				field.Set(tagsValue)
			} else {
				return fmt.Errorf("Unsupported type for key: %s", fields[0])
			}
		}
	}

	err := c.queryStruct("POST", "/teams", team, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminImportTeams(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	// Flush all existing entries
	if ctx.Bool("flush") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Flush all teams (yes/no): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		if strings.ToLower(input) != "yes" {
			return fmt.Errorf("User aborted flush operation")
		}

		teams := []api.AdminTeam{}
		err := c.queryStruct("GET", "/teams", nil, &teams)
		if err != nil {
			return err
		}

		for _, team := range teams {
			err := c.queryStruct("DELETE", fmt.Sprintf("/teams/%d", team.ID), nil, nil)
			if err != nil {
				return err
			}
		}
	}

	// Read th file
	content, err := ioutil.ReadFile(ctx.Args().Get(0))
	if err != nil {
		return err
	}

	// Parse the JSON file
	teams := []api.AdminTeam{}
	err = json.Unmarshal(content, &teams)
	if err != nil {
		return err
	}

	// Create the teams
	for _, team := range teams {
		err := c.queryStruct("POST", "/teams", team, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) cmdAdminDeleteTeam(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	err := c.queryStruct("DELETE", fmt.Sprintf("/teams/%s", ctx.Args().Get(0)), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminListTeams(ctx *cli.Context) error {
	// Get the data
	resp := []api.AdminTeam{}

	err := c.queryStruct("GET", "/teams", nil, &resp)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Country", "Website", "Subnets", "Notes", "Tags"})
	table.SetBorder(false)

	for _, entry := range resp {
		table.Append([]string{
			fmt.Sprintf("%d", entry.ID),
			entry.Name,
			entry.Country,
			entry.Website,
			entry.Subnets,
			entry.Notes,
			utils.PackTags(entry.Tags),
		})
	}

	table.Render()

	return nil
}

func (c *client) cmdAdminUpdateTeam(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	team := api.AdminTeam{}
	err := c.queryStruct("GET", fmt.Sprintf("/teams/%s", ctx.Args().Get(0)), nil, &team)
	if err != nil {
		return err
	}

	if ctx.NArg() > 1 {
		v := reflect.ValueOf(&team)

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
			} else if field.Type() == reflect.TypeOf(map[string]string{}) {
				tags, err := utils.ParseTags(fields[1])
				if err != nil {
					return err
				}

				tagsValue := reflect.ValueOf(tags)
				field.Set(tagsValue)
			} else {
				return fmt.Errorf("Unsupported type for key: %s", fields[0])
			}
		}
	}

	err = c.queryStruct("PUT", fmt.Sprintf("/teams/%s", ctx.Args().Get(0)), team.AdminTeamPut, nil)
	if err != nil {
		return err
	}

	return nil
}
