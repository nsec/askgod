package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/urfave/cli/v3"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (c *client) cmdAdminMonitorLog(_ context.Context, cmd *cli.Command) error {
	// Parse the arguments
	logLvl, err := log15.LvlFromString(cmd.String("loglevel"))
	if err != nil {
		return err
	}

	// Connection handler
	conn, err := c.websocket("/events?type=logging")
	if err != nil {
		return err
	}

	// Process the messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		event := api.Event{}

		err = json.Unmarshal(data, &event)
		if err != nil {
			continue
		}

		if event.Type != "logging" {
			continue
		}

		logEntry := api.EventLogging{}

		err = json.Unmarshal(event.Metadata, &logEntry)
		if err != nil {
			continue
		}

		lvl, err := log15.LvlFromString(logEntry.Level)
		if err != nil {
			continue
		}

		if lvl > logLvl {
			continue
		}

		ctx := []any{}
		for k, v := range logEntry.Context {
			ctx = append(ctx, k)
			ctx = append(ctx, v)
		}

		record := log15.Record{
			Time: event.Timestamp,
			Lvl:  lvl,
			Msg:  logEntry.Message,
			Ctx:  ctx,
		}

		format := log15.TerminalFormat()
		_, _ = fmt.Printf("[%s] %s", event.Server, format.Format(&record)) //nolint:forbidigo
	}

	return nil //nolint:nilerr
}

func (c *client) cmdAdminMonitorFlags(_ context.Context, cmd *cli.Command) error {
	humanOnly := cmd.Bool("human")

	// Connection handler
	conn, err := c.websocket("/events?type=flags")
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	// Process the messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		event := api.Event{}

		err = json.Unmarshal(data, &event)
		if err != nil {
			continue
		}

		if event.Type != "flags" {
			continue
		}

		score := api.EventFlag{}

		err = json.Unmarshal(event.Metadata, &score)
		if err != nil {
			continue
		}

		if humanOnly && (strings.Contains(score.Source, "agent") || strings.Contains(score.Source, "mcp")) {
			continue
		}

		team := fmt.Sprintf("id=%d", score.Team.ID)
		if score.Team.Tags["infra"] != "" {
			team = score.Team.Tags["infra"]
		}

		switch score.Type {
		case "valid":
			_, _ = fmt.Printf("[%s][%s] Team \"%s\" (%s) scored %d points with \"%s\" (id=%d) (%s) [%s]\n", //nolint:forbidigo
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, team, score.Value, score.Input, score.Flag.ID, utils.PackTags(score.Flag.Tags), score.Source)
		case "duplicate":
			_, _ = fmt.Printf("[%s][%s] Team \"%s\" (%s) re-submitted \"%s\" (id=%d) (%s) [%s]\n", //nolint:forbidigo
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, team, score.Input, score.Flag.ID, utils.PackTags(score.Flag.Tags), score.Source)
		case "invalid":
			_, _ = fmt.Printf("[%s][%s] Team \"%s\" (%s) submitted invalid flag \"%s\" [%s]\n", //nolint:forbidigo
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, team, score.Input, score.Source)
		default:
		}
	}

	return nil //nolint:nilerr
}
