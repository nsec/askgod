package main

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	c := client{}

	app := cli.NewApp()
	app.Name = "askgod"
	app.Usage = "CTF scoring system - client"
	app.HideVersion = true
	app.HideHelp = true
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "s,server",
			EnvVar:      "ASKGOD_SERVER",
			Value:       "https://askgod.nsec",
			Usage:       "URL of askgod server",
			Destination: &c.server,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "admin",
			Usage:  "Admin functions",
			Hidden: true,
			Subcommands: []cli.Command{
				{
					Name:     "config",
					Usage:    "Show the server config",
					Category: "server",
					Action:   c.cmdAdminConfig,
				},
				{
					Name:     "monitor-log",
					Usage:    "Show live log messages from the server",
					Category: "server",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "l,loglevel",
							Usage: "One of critical, error, warn, info or debug",
							Value: "info",
						},
					},
					Action: c.cmdAdminMonitorLog,
				},
				{
					Name:     "monitor-flags",
					Usage:    "Show a live stream of submitted flags",
					Category: "server",
					Action:   c.cmdAdminMonitorFlags,
				},

				{
					Name:      "add-flag",
					Usage:     "Add a new flag",
					ArgsUsage: "[key=value...]",
					Category:  "flags",
					Action:    c.cmdAdminAddFlag,
				},
				{
					Name:     "delete-flag",
					Usage:    "Delete a flag",
					Category: "flags",
					Action:   c.cmdAdminDeleteFlag,
				},
				{
					Name:      "import-flags",
					Usage:     "Import a list of flags",
					ArgsUsage: "<filename>",
					Category:  "flags",
					Action:    c.cmdAdminImportFlags,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "flush",
							Usage: "Remove all existings flags",
						},
					},
				},
				{
					Name:     "list-flags",
					Usage:    "List all the flags",
					Category: "flags",
					Action:   c.cmdAdminListFlags,
				},
				{
					Name:      "update-flag",
					Usage:     "Update a flag",
					ArgsUsage: "<id> [key=value...]",
					Category:  "flags",
					Action:    c.cmdAdminUpdateFlag,
				},

				{
					Name:      "add-team",
					Usage:     "Add a new team",
					ArgsUsage: "[key=value...]",
					Category:  "teams",
					Action:    c.cmdAdminAddTeam,
				},
				{
					Name:     "delete-team",
					Usage:    "Delete a team",
					Category: "teams",
					Action:   c.cmdAdminDeleteTeam,
				},
				{
					Name:      "import-team",
					Usage:     "Import a list of teams",
					ArgsUsage: "<filename>",
					Category:  "teams",
					Action:    c.cmdAdminImportTeams,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "flush",
							Usage: "Remove all existings teams",
						},
					},
				},
				{
					Name:     "list-teams",
					Usage:    "List all the teams",
					Category: "teams",
					Action:   c.cmdAdminListTeams,
				},
				{
					Name:      "update-team",
					Usage:     "Update a team",
					ArgsUsage: "<id> [key=value...]",
					Category:  "teams",
					Action:    c.cmdAdminUpdateTeam,
				},

				{
					Name:      "add-score",
					Usage:     "Add a new score entry",
					ArgsUsage: "[key=value...]",
					Category:  "scores",
					Action:    c.cmdAdminAddScore,
				},
				{
					Name:     "delete-score",
					Usage:    "Delete a score entry",
					Category: "scores",
					Action:   c.cmdAdminDeleteScore,
				},
				{
					Name:      "import-scores",
					Usage:     "Import a list of scores",
					ArgsUsage: "<filename>",
					Category:  "scores",
					Action:    c.cmdAdminImportScores,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "flush",
							Usage: "Remove all existings scores",
						},
					},
				},
				{
					Name:     "list-scores",
					Usage:    "List all the score entries",
					Category: "scores",
					Action:   c.cmdAdminListScores,
				},
				{
					Name:      "update-score",
					Usage:     "Update a score entry",
					ArgsUsage: "<id> [key=value...]",
					Category:  "scores",
					Action:    c.cmdAdminUpdateScore,
				},
			},
		},

		{
			Name:      "details",
			Usage:     "Query and set team details",
			ArgsUsage: "[key=value...]",
			Action:    c.cmdDetails,
		},

		{
			Name:   "history",
			Usage:  "List all submitted flags",
			Action: c.cmdHistory,
		},

		{
			Name:   "timeline",
			Usage:  "List team scores over time",
			Action: c.cmdTimeline,
			Hidden: true,
		},

		{
			Name:  "scoreboard",
			Usage: "Show the scoreboard",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l,live",
					Usage: "Keep updating the scoreboard as it changes",
				},
			},
			Action: c.cmdScoreboard,
		},

		{
			Name:      "submit",
			Usage:     "Submit a flag",
			ArgsUsage: "<flag>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "n,notes",
					Usage: "Some notes to remind you of the flag",
				},
			},
			Action: c.cmdSubmit,
		},

		{
			Name:   "status",
			Usage:  "Server and event status",
			Action: c.cmdStatus,
		},
	}

	app.Before = func(ctx *cli.Context) error {
		return c.setupClient()
	}

	app.Run(os.Args)
}
