package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (c *client) cmdAdminAddTeam(ctx *cli.Context) error {
	team := api.AdminTeamPost{}

	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args() {
			err := setStructKey(&team, arg)
			if err != nil {
				return err
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

		err := c.queryStruct("DELETE", "/teams?empty=1", nil, nil)
		if err != nil {
			return err
		}
	}

	// Read the file
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
		for _, arg := range ctx.Args()[1:] {
			err := setStructKey(&team, arg)
			if err != nil {
				return err
			}
		}
	}

	err = c.queryStruct("PUT", fmt.Sprintf("/teams/%s", ctx.Args().Get(0)), team.AdminTeamPut, nil)
	if err != nil {
		return err
	}

	return nil
}
