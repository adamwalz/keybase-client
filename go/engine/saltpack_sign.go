// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package engine

import (
	"io"

	"github.com/adamwalz/keybase-client/go/libkb"
	keybase1 "github.com/adamwalz/keybase-client/go/protocol/keybase1"
)

// SaltpackSign is an engine.
type SaltpackSign struct {
	libkb.Contextified
	arg *SaltpackSignArg
	key libkb.NaclSigningKeyPair
}

type SaltpackSignArg struct {
	Sink   io.WriteCloser
	Source io.ReadCloser
	Opts   keybase1.SaltpackSignOptions
}

// NewSaltpackSign creates a SaltpackSign engine.
func NewSaltpackSign(g *libkb.GlobalContext, arg *SaltpackSignArg) *SaltpackSign {
	return &SaltpackSign{
		arg:          arg,
		Contextified: libkb.NewContextified(g),
	}
}

// Name is the unique engine name.
func (e *SaltpackSign) Name() string {
	return "SaltpackSign"
}

// GetPrereqs returns the engine prereqs.
func (e *SaltpackSign) Prereqs() Prereqs {
	return Prereqs{
		Device: true,
	}
}

// RequiredUIs returns the required UIs.
func (e *SaltpackSign) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{
		libkb.SecretUIKind,
	}
}

// SubConsumers returns the other UI consumers for this engine.
func (e *SaltpackSign) SubConsumers() []libkb.UIConsumer {
	return nil
}

// Run starts the engine.
func (e *SaltpackSign) Run(m libkb.MetaContext) error {
	if err := e.loadKey(m); err != nil {
		return err
	}

	saltpackVersion, err := libkb.SaltpackVersionFromArg(e.arg.Opts.SaltpackVersion)
	if err != nil {
		return err
	}

	if e.arg.Opts.Detached {
		return libkb.SaltpackSignDetached(e.G(), e.arg.Source, e.arg.Sink, e.key, e.arg.Opts.Binary, saltpackVersion)
	}

	return libkb.SaltpackSign(e.G(), e.arg.Source, e.arg.Sink, e.key, e.arg.Opts.Binary, saltpackVersion)
}

func (e *SaltpackSign) loadKey(m libkb.MetaContext) error {
	loggedIn, uid, err := isLoggedInWithUIDAndError(m)
	if err != nil {
		return err
	}
	if !loggedIn {
		return libkb.NewLoginRequiredError("login required for signing")
	}
	key, err := m.G().ActiveDevice.SigningKeyWithUID(uid)
	if err != nil {
		return err
	}
	kp, ok := key.(libkb.NaclSigningKeyPair)
	if !ok || kp.Private == nil {
		return libkb.KeyCannotSignError{}
	}
	e.key = kp
	return nil
}
