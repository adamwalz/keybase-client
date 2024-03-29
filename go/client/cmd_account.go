// Copyright 2018 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"sort"

	"github.com/keybase/cli"
	"github.com/adamwalz/keybase-client/go/libcmdline"
	"github.com/adamwalz/keybase-client/go/libkb"
)

func NewCmdAccount(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	subcommands := []cli.Command{
		NewCmdAccountDelete(cl, g),
		NewCmdAccountLockdown(cl, g),
		NewCmdAccountRecoverUsername(cl, g),
		NewCmdAccountContactSettings(cl, g),
		NewCmdEmail(cl, g),
		NewCmdPhoneNumber(cl, g),
		newCmdUploadAvatar(cl, g, false /* hidden */),
		NewCmdAccountResetCancel(cl, g),
	}
	subcommands = append(subcommands, getBuildSpecificAccountCommands(cl, g)...)
	sort.Sort(cli.ByName(subcommands))
	return cli.Command{
		Name:         "account",
		Usage:        "Modify your account",
		ArgumentHelp: "[arguments...]",
		Subcommands:  subcommands,
	}
}
