package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/nsec/askgod/internal/daemon"
)

func main() {
	app := &cli.Command{}
	app.Name = "askgod-server"
	app.Usage = "CTF scoring system - server"
	app.ArgsUsage = "<config>"
	app.HideVersion = true
	app.EnableShellCompletion = true

	app.Action = func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() == 0 {
			_ = cli.ShowAppHelp(cmd)

			return errors.New("missing required arguments")
		}

		d, err := daemon.NewDaemon(cmd.Args().Get(0))
		if err != nil {
			return err
		}

		return d.Run(ctx)
	}

	err := app.Run(context.TODO(), os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		os.Exit(1)
	}
}
