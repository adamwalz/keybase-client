// Copyright 2018 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"github.com/keybase/cli"
	"github.com/adamwalz/keybase-client/go/libcmdline"
	"github.com/adamwalz/keybase-client/go/libkb"
)

func NewCmdPhoneNumber(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	return cli.Command{
		Name:     "phonenumber",
		Usage:    "Manage your phone numbers",
		Unlisted: true,
		Subcommands: []cli.Command{
			NewCmdAddPhoneNumber(cl, g),
			NewCmdEditPhoneNumber(cl, g),
			NewCmdDeletePhoneNumber(cl, g),
			NewCmdListPhoneNumbers(cl, g),
			NewCmdVerifyPhoneNumber(cl, g),
			NewCmdSetVisibilityPhoneNumber(cl, g),
		},
	}
}
