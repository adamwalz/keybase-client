// Copyright 2019 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.
//go:build darwin || netbsd || openbsd || freebsd
// +build darwin netbsd openbsd freebsd

package client

import (
	"github.com/keybase/cli"
	"github.com/adamwalz/keybase-client/go/libcmdline"
	"github.com/adamwalz/keybase-client/go/libkb"
)

func platformSpecificCtlCommands(cl *libcmdline.CommandLine, g *libkb.GlobalContext) []cli.Command {
	return nil
}
