// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"golang.org/x/net/context"

	"github.com/keybase/cli"
	"github.com/adamwalz/keybase-client/go/libcmdline"
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/protocol/keybase1"
)

type CmdInterestingPeople struct {
	libkb.Contextified
	maxUsers  int
	namespace string
}

func (c *CmdInterestingPeople) ParseArgv(ctx *cli.Context) error {
	c.maxUsers = ctx.Int("maxusers")
	c.namespace = ctx.String("namespace")
	return nil
}

func (c *CmdInterestingPeople) Run() error {
	cli, err := GetUserClient(c.G())
	if err != nil {
		return err
	}

	users, err := cli.InterestingPeople(context.Background(), keybase1.InterestingPeopleArg{MaxUsers: c.maxUsers, Namespace: c.namespace})
	if err != nil {
		return err
	}

	for _, user := range users {
		_ = c.G().UI.GetTerminalUI().Output(user.Username + "\n")
	}

	return nil
}
func NewCmdInterestingPeopleRunner(g *libkb.GlobalContext) *CmdInterestingPeople {
	return &CmdInterestingPeople{Contextified: libkb.NewContextified(g)}
}

func NewCmdInterestingPeople(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	ret := cli.Command{
		Name:        "interesting-people",
		Description: "List interesting people that you might want to interact with",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "maxusers",
				Usage: "Max users to return",
				Value: 20,
			},
			cli.StringFlag{
				Name:  "namespace",
				Usage: "Namespace to filter recommendations for",
				Value: "people",
			},
		},
		Action: func(c *cli.Context) {
			cl.ChooseCommand(NewCmdInterestingPeopleRunner(g), "interesting-people", c)
		},
	}
	return ret
}

func (c *CmdInterestingPeople) GetUsage() libkb.Usage {
	return libkb.Usage{
		Config:    true,
		KbKeyring: true,
		API:       true,
	}
}
