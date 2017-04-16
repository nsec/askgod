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

func (c *client) cmdAdminAddFlag(ctx *cli.Context) error {
	flag := api.AdminFlagPost{}

	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args() {
			err := setStructKey(&flag, arg)
			if err != nil {
				return err
			}
		}
	}

	err := c.queryStruct("POST", "/flags", flag, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminDeleteFlag(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	err := c.queryStruct("DELETE", fmt.Sprintf("/flags/%s", ctx.Args().Get(0)), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminImportFlags(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	// Flush all existing entries
	if ctx.Bool("flush") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Flush all flags (yes/no): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		if strings.ToLower(input) != "yes" {
			return fmt.Errorf("User aborted flush operation")
		}

		flags := []api.AdminFlag{}
		err := c.queryStruct("GET", "/flags", nil, &flags)
		if err != nil {
			return err
		}

		for _, flag := range flags {
			err := c.queryStruct("DELETE", fmt.Sprintf("/flags/%d", flag.ID), nil, nil)
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
	flags := []api.AdminFlag{}
	err = json.Unmarshal(content, &flags)
	if err != nil {
		return err
	}

	// Create the flags
	for _, flag := range flags {
		err := c.queryStruct("POST", "/flags", flag, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) cmdAdminListFlags(ctx *cli.Context) error {
	// Get the data
	resp := []api.AdminFlag{}

	err := c.queryStruct("GET", "/flags", nil, &resp)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Flag", "Value", "Return string", "Description", "Tags"})
	table.SetBorder(false)

	for _, entry := range resp {
		table.Append([]string{
			fmt.Sprintf("%d", entry.ID),
			entry.Flag,
			fmt.Sprintf("%d", entry.Value),
			entry.ReturnString,
			entry.Description,
			utils.PackTags(entry.Tags),
		})
	}

	table.Render()

	return nil
}

func (c *client) cmdAdminUpdateFlag(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	flag := api.AdminFlag{}
	err := c.queryStruct("GET", fmt.Sprintf("/flags/%s", ctx.Args().Get(0)), nil, &flag)
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
	}

	err = c.queryStruct("PUT", fmt.Sprintf("/flags/%s", ctx.Args().Get(0)), flag.AdminFlagPut, nil)
	if err != nil {
		return err
	}

	return nil
}
