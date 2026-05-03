package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdAdminAddScore(ctx context.Context, cmd *cli.Command) error {
	score := api.AdminScorePost{}

	if cmd.NArg() > 0 {
		for _, arg := range cmd.Args().Slice() {
			err := setStructKey(&score, arg)
			if err != nil {
				return err
			}
		}
	}

	err := c.queryStruct(ctx, "POST", "/scores", score, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminDeleteScore(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		_ = cli.ShowSubcommandHelp(cmd)

		return nil
	}

	err := c.queryStruct(ctx, "DELETE", "/scores/"+cmd.Args().Get(0), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminImportScores(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		_ = cli.ShowSubcommandHelp(cmd)

		return nil
	}

	// Flush all existing entries
	if cmd.Bool("flush") {
		reader := bufio.NewReader(os.Stdin)
		_, _ = fmt.Print("Flush all scores (yes/no): ") //nolint:forbidigo
		input, _ := reader.ReadString('\n')

		input = strings.TrimSuffix(input, "\n")
		if strings.TrimSpace(strings.ToLower(input)) != "yes" {
			return errors.New("user aborted flush operation")
		}

		err := c.queryStruct(ctx, "DELETE", "/scores?empty=1", nil, nil)
		if err != nil {
			return err
		}
	}

	// Read the file
	content, err := os.ReadFile(cmd.Args().Get(0))
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
	err = c.queryStruct(ctx, "POST", "/scores?bulk=1", scores, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) cmdAdminListScores(ctx context.Context, _ *cli.Command) error {
	// Get the data
	resp := []api.AdminScore{}

	err := c.queryStruct(ctx, "GET", "/scores", nil, &resp)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "TeamID", "FlagID", "Value", "Submit time", "Source", "Notes"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, entry := range resp {
		table.Append([]string{
			strconv.FormatInt(entry.ID, 10),
			strconv.FormatInt(entry.TeamID, 10),
			strconv.FormatInt(entry.FlagID, 10),
			strconv.FormatInt(entry.Value, 10),
			entry.SubmitTime.Local().Format(layout),
			entry.Source,
			entry.Notes,
		})
	}

	table.Render()

	return nil
}

func (c *client) cmdAdminUpdateScore(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		_ = cli.ShowSubcommandHelp(cmd)

		return nil
	}

	score := api.AdminScore{}

	err := c.queryStruct(ctx, "GET", "/scores/"+cmd.Args().Get(0), nil, &score)
	if err != nil {
		return err
	}

	if cmd.NArg() > 1 {
		for _, arg := range cmd.Args().Slice()[1:] {
			err := setStructKey(&score, arg)
			if err != nil {
				return err
			}
		}
	}

	err = c.queryStruct(ctx, "PUT", "/scores/"+cmd.Args().Get(0), score.AdminScorePut, nil)
	if err != nil {
		return err
	}

	return nil
}
