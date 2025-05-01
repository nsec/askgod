package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdAdminConfig(ctx *cli.Context) error {
	// Get the data
	resp := api.Config{}

	err := c.queryStruct("GET", "/config", nil, &resp)
	if err != nil {
		return err
	}

	// Process any field update
	if ctx.NArg() > 0 {
		for _, arg := range ctx.Args().Slice() {
			err := setStructKey(&resp, arg)
			if err != nil {
				return err
			}
		}

		// Update the team
		err = c.queryStruct("PUT", "/config", resp.ConfigPut, nil)
		if err != nil {
			return err
		}

		return nil
	}

	data, err := yaml.Marshal(&resp)
	if err != nil {
		return err
	}

	_, _ = fmt.Printf("%s", data)

	return nil
}
