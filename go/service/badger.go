// Copyright 2017 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"github.com/adamwalz/keybase-client/go/badges"
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/protocol/keybase1"
	"github.com/keybase/go-framed-msgpack-rpc/rpc"
	"golang.org/x/net/context"
)

type badgerHandler struct {
	*BaseHandler
	libkb.Contextified

	badger *badges.Badger
}

func newBadgerHandler(xp rpc.Transporter, g *libkb.GlobalContext, badger *badges.Badger) *badgerHandler {
	return &badgerHandler{
		BaseHandler:  NewBaseHandler(g, xp),
		Contextified: libkb.NewContextified(g),
		badger:       badger,
	}
}

func (a *badgerHandler) GetBadgeState(ctx context.Context) (res keybase1.BadgeState, err error) {
	a.G().Trace("GetBadgeState", &err)()
	return a.badger.State().Export(ctx)
}
