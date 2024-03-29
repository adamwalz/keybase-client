// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package engine

import (
	"fmt"
	"io"

	"github.com/adamwalz/keybase-client/go/libkb"
	keybase1 "github.com/adamwalz/keybase-client/go/protocol/keybase1"

	"github.com/keybase/go-crypto/openpgp"
	"github.com/keybase/go-crypto/openpgp/clearsign"
)

type PGPSignEngine struct {
	arg *PGPSignArg
	libkb.Contextified

	warnings libkb.HashSecurityWarnings
}

type PGPSignArg struct {
	Sink   io.WriteCloser
	Source io.ReadCloser
	Opts   keybase1.PGPSignOptions
}

func (p *PGPSignEngine) Prereqs() Prereqs {
	return Prereqs{
		Device: true,
	}
}

func (p *PGPSignEngine) Name() string {
	return "PGPSignEngine"
}

func (p *PGPSignEngine) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{
		libkb.SecretUIKind,
		libkb.PgpUIKind,
	}
}

func (p *PGPSignEngine) SubConsumers() []libkb.UIConsumer {
	return nil
}

func NewPGPSignEngine(g *libkb.GlobalContext, arg *PGPSignArg) *PGPSignEngine {
	return &PGPSignEngine{
		arg:          arg,
		Contextified: libkb.NewContextified(g),
	}
}

func (p *PGPSignEngine) Run(m libkb.MetaContext) (err error) {
	var key libkb.GenericKey
	var pgp *libkb.PGPKeyBundle
	var ok bool
	var dumpTo io.WriteCloser
	var written int64

	defer func() {
		if dumpTo != nil {
			if e := dumpTo.Close(); e != nil {
				p.G().Log.Warning("error closing dumpTo: %s", e)
			}
		}
		if e := p.arg.Sink.Close(); e != nil {
			p.G().Log.Warning("error closing Sink: %s", e)
		}
		if e := p.arg.Source.Close(); e != nil {
			p.G().Log.Warning("error closing Source: %s", e)
		}
	}()

	me, err := libkb.LoadMe(libkb.NewLoadUserArg(p.G()))
	if err != nil {
		return err
	}

	ska := libkb.SecretKeyArg{
		Me:       me,
		KeyType:  libkb.PGPKeyType,
		KeyQuery: p.arg.Opts.KeyQuery,
	}
	key, err = p.G().Keyrings.GetSecretKeyWithPrompt(m, m.SecretKeyPromptArg(ska, "command-line signature"))
	if err != nil {
		return
	} else if pgp, ok = key.(*libkb.PGPKeyBundle); !ok {
		err = fmt.Errorf("Can only sign with PGP keys")
		return
	}

	p.warnings = libkb.HashSecurityWarnings{}
	if w := pgp.SecurityWarnings(
		libkb.HashSecurityWarningOurIdentityHash,
	); len(w) > 0 {
		p.warnings = append(p.warnings, w...)
	}
	for _, warning := range p.warnings.Strings() {
		if err := m.UIs().PgpUI.OutputPGPWarning(m.Ctx(), keybase1.OutputPGPWarningArg{
			Warning: warning,
		}); err != nil {
			return err
		}
	}

	bo := p.arg.Opts.BinaryOut
	bi := p.arg.Opts.BinaryIn
	pgpe := pgp.Entity
	mode := p.arg.Opts.Mode

	switch mode {
	case keybase1.SignMode_ATTACHED:
		dumpTo, err = libkb.AttachedSignWrapper(p.arg.Sink, *pgp, !bo)
	case keybase1.SignMode_DETACHED:
		switch {
		case bi && bo:
			err = openpgp.DetachSign(p.arg.Sink, pgpe, p.arg.Source, nil)
		case bi && !bo:
			err = openpgp.ArmoredDetachSign(p.arg.Sink, pgpe, p.arg.Source, nil)
		case !bi && bo:
			err = openpgp.DetachSignText(p.arg.Sink, pgpe, p.arg.Source, nil)
		default:
			err = openpgp.ArmoredDetachSignText(p.arg.Sink, pgpe, p.arg.Source, nil)
		}
	case keybase1.SignMode_CLEAR:
		dumpTo, err = clearsign.Encode(p.arg.Sink, pgp.PrivateKey, nil)
	default:
		err = fmt.Errorf("unrecognized sign mode: %d", int(mode))
	}

	if err != nil {
		return
	}

	if dumpTo != nil {
		written, err = io.Copy(dumpTo, p.arg.Source)
		if err == nil && written == 0 {
			p.G().Log.Debug("Empty source file.")
		}
	}
	return
}
