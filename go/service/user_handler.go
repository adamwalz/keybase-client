// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/adamwalz/keybase-client/go/gregor"
	"github.com/adamwalz/keybase-client/go/libkb"
	gregor1 "github.com/adamwalz/keybase-client/go/protocol/gregor1"
	keybase1 "github.com/adamwalz/keybase-client/go/protocol/keybase1"
)

const userHandlerName = "userHandler"

type userHandler struct {
	libkb.Contextified
	userBlockedHandlers []UserBlockedHandler
}

type UserBlockedHandler interface {
	UserBlocked(m libkb.MetaContext, badUIDs map[keybase1.UID]bool) error
}

func newUserHandler(g *libkb.GlobalContext) *userHandler {
	return &userHandler{
		Contextified: libkb.NewContextified(g),
	}
}

func (r *userHandler) PushUserBlockedHandler(h UserBlockedHandler) {
	r.userBlockedHandlers = append(r.userBlockedHandlers, h)
}

func (r *userHandler) Create(ctx context.Context, cli gregor1.IncomingInterface, category string, item gregor.Item) (bool, error) {
	m := libkb.NewMetaContext(ctx, r.G())
	switch category {
	case "user.key_change":
		return true, r.keyChange(m)
	case "user.identity_change":
		return true, r.identityChange(m)
	case "user.password_change":
		return true, r.passwordChange(m, cli, category, item)
	case "user.passphrase_state":
		return true, r.passphraseStateUpdate(m, cli, category, item)
	case "user.blocked":
		return true, r.userBlocked(m, cli, category, item)
	default:
		if strings.HasPrefix(category, "user.") {
			return false, fmt.Errorf("unknown userHandler category: %q", category)
		}
		return false, nil
	}
}

func (r *userHandler) keyChange(m libkb.MetaContext) error {
	m.G().KeyfamilyChanged(m.Ctx(), m.G().Env.GetUID())

	// check if this device was just revoked and if so, log out
	return m.LogoutAndDeprovisionIfRevoked()
}

func (r *userHandler) identityChange(m libkb.MetaContext) error {
	m.G().UserChanged(m.Ctx(), m.G().Env.GetUID())
	return nil
}

func (r *userHandler) passwordChange(m libkb.MetaContext, cli gregor1.IncomingInterface, category string, item gregor.Item) error {
	m.Debug("userHandler: %s received", category)
	return r.G().GregorState.DismissItem(m.Ctx(), cli, item.Metadata().MsgID())
}

func (r *userHandler) passphraseStateUpdate(m libkb.MetaContext, cli gregor1.IncomingInterface, category string, item gregor.Item) error {
	m.Debug("userHandler: %s received", category)
	var msg keybase1.UserPassphraseStateMsg
	if err := json.Unmarshal(item.Body().Bytes(), &msg); err != nil {
		m.Warning("error unmarshaling user.passphrase_update item: %s", err)
		return err
	}
	libkb.MaybeSavePassphraseState(m, msg.PassphraseState)
	r.G().NotifyRouter.HandlePasswordChanged(m.Ctx(), msg.PassphraseState)
	// Don't dismiss the item, so other devices know about it
	return nil
}

func (r *userHandler) userBlocked(m libkb.MetaContext, cli gregor1.IncomingInterface, category string, item gregor.Item) error {
	m.Debug("userHandler: %s received", category)
	var msg keybase1.UserBlockedBody
	if err := json.Unmarshal(item.Body().Bytes(), &msg); err != nil {
		m.Warning("error unmarshaling user.blocked item: %s", err)
		return err
	}
	m.Debug("Got user.blocked item: %+v", msg)
	badUIDs := make(map[keybase1.UID]bool)

	for _, r := range msg.Blocks {
		if (r.Chat != nil && *r.Chat) || (r.Follow != nil && *r.Follow) {
			badUIDs[r.Uid] = true
		}

		// regardless of status, clear user card cache for any uid
		// that changed
		if err := m.G().CardCache().Delete(r.Uid); err != nil {
			m.Debug("CardCache.Delete(%s) error: %s", r.Uid, err)
		}
	}

	// Ignore the error if we fail to block properly
	_ = libkb.NewServertrustTrackerSyncer(m.G(), msg.Uid, libkb.FollowDirectionFollowing).Block(m, badUIDs)

	m.Debug("Got user.blocked blocked UIDs %+v", badUIDs)
	for _, h := range r.userBlockedHandlers {
		tmp := h.UserBlocked(m, badUIDs)
		if tmp != nil {
			m.Warning("Error handling UserBlocked message: %s", tmp)
		}
	}

	r.G().NotifyRouter.HandleUserBlocked(m.Ctx(), msg)

	r.G().GetStellar().Refresh(m, "user.blocked message")

	// dismiss this locally so it is only processed once per device
	if err := r.G().GregorState.LocalDismissItem(m.Ctx(), item.Metadata().MsgID()); err != nil {
		m.Warning("error in LocalDismissItem: %s", err)
	}

	return nil
}

func (r *userHandler) Dismiss(ctx context.Context, cli gregor1.IncomingInterface, category string, item gregor.Item) (bool, error) {
	return false, nil
}

func (r *userHandler) IsAlive() bool {
	return true
}

func (r *userHandler) Name() string {
	return userHandlerName
}
