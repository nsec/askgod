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

	data, err := yaml.Marshal(&resp)
	if err != nil {
		return err
	}

	fmt.Printf("%s", data)

	return nil
}
