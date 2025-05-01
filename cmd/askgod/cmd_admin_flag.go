package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (c *client) cmdAdminAddFlag(ctx *cli.Context) error {
	flag := api.AdminFlagPost{}

	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args().Slice() {
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
		_ = cli.ShowSubcommandHelp(ctx)

		return nil
	}

	err := c.queryStruct("DELETE", "/flags/"+ctx.Args().Get(0), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminImportFlags(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		_ = cli.ShowSubcommandHelp(ctx)

		return nil
	}

	// Flush all existing entries
	if ctx.Bool("flush") {
		reader := bufio.NewReader(os.Stdin)
		_, _ = fmt.Printf("Flush all flags (yes/no): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		if strings.TrimSpace(strings.ToLower(input)) != "yes" {
			return errors.New("user aborted flush operation")
		}

		err := c.queryStruct("DELETE", "/flags?empty=1", nil, nil)
		if err != nil {
			return err
		}
	}

	// Read the file
	content, err := os.ReadFile(ctx.Args().Get(0))
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
	err = c.queryStruct("POST", "/flags?bulk=1", flags, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminListFlags(_ *cli.Context) error {
	// Get the data
	resp := []api.AdminFlag{}

	err := c.queryStruct("GET", "/flags", nil, &resp)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Flag", "Value", "Return string", "Description", "Tags"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, entry := range resp {
		table.Append([]string{
			strconv.FormatInt(entry.ID, 10),
			entry.Flag,
			strconv.FormatInt(entry.Value, 10),
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
		_ = cli.ShowSubcommandHelp(ctx)

		return nil
	}

	flag := api.AdminFlag{}
	err := c.queryStruct("GET", "/flags/"+ctx.Args().Get(0), nil, &flag)
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
	}

	err = c.queryStruct("PUT", "/flags/"+ctx.Args().Get(0), flag.AdminFlagPut, nil)
	if err != nil {
		return err
	}

	return nil
}
