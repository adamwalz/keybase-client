package storage

import (
	"context"
	"fmt"

	"github.com/adamwalz/keybase-client/go/chat/globals"
	"github.com/adamwalz/keybase-client/go/chat/utils"
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/protocol/chat1"
	"github.com/adamwalz/keybase-client/go/protocol/keybase1"
)

type breakTracker struct {
	globals.Contextified
	utils.DebugLabeler
}

func newBreakTracker(g *globals.Context) *breakTracker {
	return &breakTracker{
		Contextified: globals.NewContextified(g),
		DebugLabeler: utils.NewDebugLabeler(g.ExternalG(), "BreakTracker", false),
	}
}

func (b *breakTracker) makeDbKey(tlfID chat1.TLFID) libkb.DbKey {
	return libkb.DbKey{
		Typ: libkb.DBChatBlocks,
		Key: fmt.Sprintf("breaks:%s", tlfID),
	}
}

func (b *breakTracker) UpdateTLF(ctx context.Context, tlfID chat1.TLFID,
	breaks []keybase1.TLFIdentifyFailure) (err error) {
	defer b.Trace(ctx, &err, "UpdateTLF(%s)", tlfID)()
	key := b.makeDbKey(tlfID)

	dat, err := encode(breaks)
	if err != nil {
		return NewInternalError(ctx, b.DebugLabeler, "encode error: %s", err.Error())
	}
	if err = b.G().LocalChatDb.PutRaw(key, dat); err != nil {
		return NewInternalError(ctx, b.DebugLabeler, "PutRaw error: %s", err.Error())
	}

	return nil
}

func (b *breakTracker) IsTLFBroken(ctx context.Context, tlfID chat1.TLFID) (res bool, err error) {
	defer b.Trace(ctx, &err, "IsTLFBroken(%s)", tlfID)()
	key := b.makeDbKey(tlfID)
	raw, found, err := b.G().LocalChatDb.GetRaw(key)
	if err != nil {
		return true, NewInternalError(ctx, b.DebugLabeler, "GetRaw error: %s", err.Error())
	}

	// Assume to be broken if we have no record
	if !found {
		return true, nil
	}

	var breaks []keybase1.TLFIdentifyFailure
	if err = decode(raw, &breaks); err != nil {
		return true, NewInternalError(ctx, b.DebugLabeler, "decode error: %s", err.Error())
	}

	return len(breaks) != 0, nil
}
