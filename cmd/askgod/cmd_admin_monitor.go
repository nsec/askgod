package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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
	flagIDs := cmd.Int64Slice("flags")

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

		if !flagMatchesIDs(score.Flag, flagIDs) {
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

// wsChannelForEventType maps user-facing event type names to websocket channel
// names. "ratelimit" events travel on the "flags" channel.
var wsChannelForEventType = map[string]string{
	"flags":     "flags",
	"logging":   "logging",
	"timeline":  "timeline",
	"ratelimit": "flags",
}

// cmdAdminMonitorHook connects to the requested event channels and calls
// <script> for every received event, passing the raw event JSON on stdin.
// The event JSON has the shape: {"server":..., "type":..., "timestamp":..., "metadata":{...}}
func (c *client) cmdAdminMonitorHook(_ context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("missing required argument: <script>")
	}

	script := cmd.Args().First()

	// Build the deduplicated list of websocket channels to subscribe to.
	seen := map[string]bool{}
	wsChannels := []string{}

	for _, t := range strings.Split(cmd.String("event"), ",") {
		t = strings.TrimSpace(t)

		ch, ok := wsChannelForEventType[t]
		if !ok {
			return fmt.Errorf("unknown event type %q (valid: flags, logging, timeline, ratelimit)", t)
		}

		if !seen[ch] {
			seen[ch] = true
			wsChannels = append(wsChannels, ch)
		}
	}

	conn, err := c.websocket("/events?type=" + strings.Join(wsChannels, ","))
	if err != nil {
		return err
	}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		hookCmd := exec.Command(script) //nolint:gosec
		hookCmd.Stdin = bytes.NewReader(data)

		if err := hookCmd.Run(); err != nil {
			_, _ = fmt.Printf("hook error: %v\n", err) //nolint:forbidigo
		}
	}

	return nil //nolint:nilerr
}
