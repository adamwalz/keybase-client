// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

//go:build production
// +build production

package saltpackkeys

import "github.com/adamwalz/keybase-client/go/libkb"

func NewRecipientKeyfinderEngineHook(getKBFSKeyfinderForTesting bool) func(arg libkb.SaltpackRecipientKeyfinderArg) libkb.SaltpackRecipientKeyfinderEngineInterface {
	if getKBFSKeyfinderForTesting {
		panic("NewRecipientKeyfinderEngineHook: getKBFSKeyfinderForTesting is true in production")
	}
	return NewSaltpackRecipientKeyfinderEngineAsInterface
}
