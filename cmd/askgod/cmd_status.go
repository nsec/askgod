package main

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
)

func (c *client) cmdStatus(_ *cli.Context) error {
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

	_, _ = fmt.Printf("%s", data)

	return nil
}
