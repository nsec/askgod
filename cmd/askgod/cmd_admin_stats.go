package main

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func isAISource(source string) bool {
	switch source {
	case api.SourceCLIAgent, api.SourceWebAgent, api.SourceMCP:
		return true
	default:
		return false
	}
}

func (c *client) cmdAdminStats(ctx context.Context, _ *cli.Command) error {
	scores := []api.AdminScore{}

	err := c.queryStruct(ctx, "GET", "/scores", nil, &scores)
	if err != nil {
		return err
	}

	teams := []api.AdminTeam{}

	err = c.queryStruct(ctx, "GET", "/teams", nil, &teams)
	if err != nil {
		return err
	}

	flags := []api.AdminFlag{}

	err = c.queryStruct(ctx, "GET", "/flags", nil, &flags)
	if err != nil {
		return err
	}

	slices.SortFunc(flags, func(a, b api.AdminFlag) int {
		return cmp.Compare(a.ID, b.ID)
	})

	const layout = "2006/01/02 15:04"

	teamCount := len(teams)

	type flagStats struct {
		firstBlood time.Time
		lastSolve  time.Time
		hasSolve   bool
		teams      map[int64]struct{}
		solveCount int
		aiCount    int
	}

	stats := make(map[int64]*flagStats, len(flags))
	for _, flag := range flags {
		stats[flag.ID] = &flagStats{teams: make(map[int64]struct{})}
	}

	for _, score := range scores {
		fs, ok := stats[score.FlagID]
		if !ok {
			continue
		}

		t := score.SubmitTime
		if !fs.hasSolve || t.Before(fs.firstBlood) {
			fs.firstBlood = t
		}

		if !fs.hasSolve || t.After(fs.lastSolve) {
			fs.lastSolve = t
		}

		fs.hasSolve = true
		fs.teams[score.TeamID] = struct{}{}
		fs.solveCount++

		if isAISource(score.Source) {
			fs.aiCount++
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"FlagID",
		"Value",
		"First blood",
		"Last solve",
		"Solve %",
		"% of agent solves",
		"Tags",
	})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, flag := range flags {
		fs := stats[flag.ID]

		firstBlood := "-"
		lastSolve := "-"
		solvePct := "0%"
		aiPct := "-"

		if fs.hasSolve {
			firstBlood = fs.firstBlood.Local().Format(layout)
			lastSolve = fs.lastSolve.Local().Format(layout)

			if teamCount > 0 {
				solvePct = fmt.Sprintf("%.1f%%", float64(len(fs.teams))/float64(teamCount)*100)
			}

			if fs.solveCount > 0 {
				aiPct = fmt.Sprintf("%.1f%%", float64(fs.aiCount)/float64(fs.solveCount)*100)
			}
		}

		table.Append([]string{
			strconv.FormatInt(flag.ID, 10),
			strconv.FormatInt(flag.Value, 10),
			firstBlood,
			lastSolve,
			solvePct,
			aiPct,
			utils.PackTags(flag.Tags),
		})
	}

	table.Render()

	return nil
}
