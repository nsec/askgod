package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// Sorting.
type byPointsAndLastSubmitTime []api.ScoreboardEntry

func (a byPointsAndLastSubmitTime) Len() int {
	return len(a)
}

func (a byPointsAndLastSubmitTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byPointsAndLastSubmitTime) Less(i, j int) bool {
	if a[i].Value != a[j].Value {
		return a[i].Value > a[j].Value
	}

	return a[i].LastSubmitTime.Before(a[j].LastSubmitTime)
}

func (c *client) cmdScoreboard(ctx *cli.Context) error {
	board := []api.ScoreboardEntry{}

	const layout = "2006/01/02 15:04"

	drawTable := func(board []api.ScoreboardEntry) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Ranking", "Team", "Points", "Last submit"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		rank := 1
		for _, entry := range board {
			lastSubmitTime := "never"
			if !entry.LastSubmitTime.IsZero() {
				lastSubmitTime = entry.LastSubmitTime.Local().Format(layout)
			}

			table.Append([]string{
				strconv.Itoa(rank),
				fmt.Sprintf("<%s> %s ", entry.Team.Country, entry.Team.Name),
				strconv.FormatInt(entry.Value, 10),
				lastSubmitTime,
			})

			rank++
		}

		table.Render()
	}

	if !ctx.Bool("live") {
		// Get the data
		err := c.queryStruct("GET", "/scoreboard", nil, &board)
		if err != nil {
			return err
		}

		sort.Sort(byPointsAndLastSubmitTime(board))

		drawTable(board)

		return nil
	}

	// Setup websocket connection
	chReady := make(chan bool, 1)
	chUpdate := make(chan bool, 1)

	conn, err := c.websocket("/events?type=timeline")
	if err != nil {
		return err
	}

	go func() {
		<-chReady
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				close(chUpdate)

				break
			}

			event := api.Event{}
			err = json.Unmarshal(data, &event)
			if err != nil {
				continue
			}

			entry := api.EventTimeline{}
			err = json.Unmarshal(event.Metadata, &entry)
			if err != nil {
				continue
			}

			// Ignore events we don't care about
			if !utils.StringInSlice(entry.Type, []string{"reload", "team-updated", "team-removed", "score-updated"}) {
				continue
			}

			// Server requests a reload of the data
			if entry.Type == "reload" {
				// Get a new dump
				board = []api.ScoreboardEntry{}
				err = c.queryStruct("GET", "/scoreboard", nil, &board)
				if err != nil {
					close(chUpdate)

					break
				}
			}

			// Try to find the line
			found := false
			for i, line := range board {
				if line.Team.ID != entry.TeamID {
					continue
				}

				// Update an existing
				found = true

				// Team is completely gone
				if entry.Type == "team-removed" {
					copy(board[i:], board[i+1:])
					board = board[:len(board)-1]

					break
				}

				// Team may have changed
				if entry.Team != nil {
					board[i].Team = api.Team{TeamPut: *entry.Team, ID: entry.TeamID}
				}

				// Score may have changed
				if entry.Score != nil {
					board[i].Value = entry.Score.Total
					board[i].LastSubmitTime = event.Timestamp
				}

				found = true

				break
			}

			// Add a new line
			if !found && entry.Team != nil {
				newEntry := api.ScoreboardEntry{
					Team:           api.Team{TeamPut: *entry.Team, ID: entry.TeamID},
					LastSubmitTime: event.Timestamp,
				}

				if entry.Score != nil {
					newEntry.Value = entry.Score.Total
				}

				board = append(board, newEntry)
			}

			// Sort the updated board ourselves
			sort.Sort(byPointsAndLastSubmitTime(board))

			chUpdate <- true
		}
	}()

	// Get the initial data
	err = c.queryStruct("GET", "/scoreboard", nil, &board)
	if err != nil {
		return err
	}

	// Ready to get websocket events
	close(chReady)

	// Refresh loop
	for {
		_, _ = fmt.Print("\033[H\033[2J")
		drawTable(board)

		// Wait for an update
		_, ok := <-chUpdate
		if !ok {
			break
		}
	}

	return nil
}
