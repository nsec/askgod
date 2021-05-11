package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// Sorting
type byPointsAndID []api.ScoreboardEntry

func (a byPointsAndID) Len() int {
	return len(a)
}

func (a byPointsAndID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byPointsAndID) Less(i, j int) bool {
	if a[i].Value != a[j].Value {
		return a[i].Value > a[j].Value
	}

	return a[i].Team.ID < a[j].Team.ID
}

func (c *client) cmdScoreboard(ctx *cli.Context) error {
	board := []api.ScoreboardEntry{}

	const layout = "2006/01/02 15:04"

	drawTable := func(board []api.ScoreboardEntry) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Team", "Points", "Last submit"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		for _, entry := range board {
			lastSubmitTime := "never"
			if !entry.LastSubmitTime.IsZero() {
				lastSubmitTime = entry.LastSubmitTime.Local().Format(layout)
			}

			table.Append([]string{
				fmt.Sprintf("<%s> %s ", entry.Team.Country, entry.Team.Name),
				fmt.Sprintf("%d", entry.Value),
				lastSubmitTime,
			})
		}

		table.Render()
	}

	if ctx.Bool("live") {
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

					// Update an existing existing
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
				sort.Sort(byPointsAndID(board))

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
			fmt.Print("\033[H\033[2J")
			drawTable(board)

			// Wait for an update
			_, ok := <-chUpdate
			if !ok {
				break
			}
		}
	} else {
		// Get the data
		err := c.queryStruct("GET", "/scoreboard", nil, &board)
		if err != nil {
			return err
		}

		drawTable(board)
	}

	return nil
}
