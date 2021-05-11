package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdStatus(ctx *cli.Context) error {
	// Get the data
	resp := api.Status{}

	err := c.queryStruct("GET", "", nil, &resp)
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
