// Copyright 2019 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

//go:build dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build dragonfly freebsd linux netbsd openbsd solaris

package status

import (
	"strings"

	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/protocol/keybase1"
)

func osSpecific(mctx libkb.MetaContext, status *keybase1.FullStatus) error {
	// TODO: on darwin, install.KeybaseServiceStatus() is implemented to get pid for service and kbfs.
	// This is currently the best way to determine if KBFS is running, so other OS's should implement
	// it.
	productVersion, buildVersion, err := libkb.OSVersionAndBuild()
	if err != nil {
		mctx.Debug("Error determining OS version: %s", err)
	}
	status.ExtStatus.PlatformInfo.OsVersion = strings.Join([]string{productVersion, buildVersion}, "-")
	return nil
}
