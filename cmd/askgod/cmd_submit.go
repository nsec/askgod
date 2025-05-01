package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdSubmit(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		_ = cli.ShowCommandHelp(ctx, "submit")

		return nil
	}

	// Prepare the input
	flag := api.FlagPost{}
	flag.Flag = ctx.Args().Get(0)
	flag.Notes = ctx.String("notes")

	// Send the flag
	resp := api.Flag{}

	err := c.queryStruct("POST", "/team/flags", flag, &resp)
	if err != nil {
		return err
	}

	// Process the points
	switch {
	case resp.Value < 0:
		_, _ = fmt.Printf("You shouldn't have sent that! You just lost your team %d points.\n", resp.Value*-1)

	case resp.Value == 0:
		_, _ = fmt.Printf("You sent a valid flag, but no points have been granted.\n")

	default:
		_, _ = fmt.Printf("Congratulations, you score your team %d points!\n", resp.Value)
	}

	// And show any message we received
	if resp.ReturnString != "" {
		_, _ = fmt.Printf("Message: %s\n", resp.ReturnString)
	}

	return nil
}
