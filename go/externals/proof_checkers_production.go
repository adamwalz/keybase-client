// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

//go:build production || staging
// +build production staging

package externals

import libkb "github.com/adamwalz/keybase-client/go/libkb"

const useDevelProofCheckers = false

func getBuildSpecificStaticProofServices() []libkb.ServiceType {
	return nil
}
