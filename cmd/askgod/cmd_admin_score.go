package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdAdminAddScore(ctx *cli.Context) error {
	score := api.AdminScorePost{}

	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args().Slice() {
			err := setStructKey(&score, arg)
			if err != nil {
				return err
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

func (c *client) cmdAdminImportScores(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	// Flush all existing entries
	if ctx.Bool("flush") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Flush all scores (yes/no): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		if strings.TrimSpace(strings.ToLower(input)) != "yes" {
			return fmt.Errorf("User aborted flush operation")
		}

		err := c.queryStruct("DELETE", "/scores?empty=1", nil, nil)
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
	scores := []api.AdminScore{}
	err = json.Unmarshal(content, &scores)
	if err != nil {
		return err
	}

	// Create the scores
	err = c.queryStruct("POST", "/scores?bulk=1", scores, nil)
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
	table.SetAutoWrapText(false)

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
		for _, arg := range ctx.Args().Slice()[1:] {
			err := setStructKey(&score, arg)
			if err != nil {
				return err
			}
		}
	}

	err = c.queryStruct("PUT", fmt.Sprintf("/scores/%s", ctx.Args().Get(0)), score.AdminScorePut, nil)
	if err != nil {
		return err
	}

	return nil
}
