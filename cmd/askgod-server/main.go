package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/internal/daemon"
)

func main() {
	app := cli.NewApp()
	app.Name = "askgod-server"
	app.Usage = "CTF scoring system - server"
	app.ArgsUsage = "<config>"
	app.HideVersion = true
	app.HideHelp = true
	app.EnableBashCompletion = true

	app.Action = func(c *cli.Context) error {
		if c.NArg() == 0 {
			_ = cli.ShowAppHelp(c)

			return errors.New("missing required arguments")
		}

		d, err := daemon.NewDaemon(c.Args().Get(0))
		if err != nil {
			return err
		}

		return d.Run()
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
