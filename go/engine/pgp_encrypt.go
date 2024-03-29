// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package engine

import (
	"errors"
	"io"

	"github.com/adamwalz/keybase-client/go/libkb"
	keybase1 "github.com/adamwalz/keybase-client/go/protocol/keybase1"
	"github.com/keybase/go-crypto/openpgp/armor"
)

type PGPEncryptArg struct {
	Recips       []string // user assertions
	Source       io.Reader
	Sink         io.WriteCloser
	NoSign       bool
	NoSelf       bool
	BinaryOutput bool
	KeyQuery     string
}

// PGPEncrypt encrypts data read from a source into a sink
// for a set of users.  It will track them if necessary.
type PGPEncrypt struct {
	arg      *PGPEncryptArg
	me       *libkb.User
	warnings libkb.HashSecurityWarnings
	libkb.Contextified
}

// NewPGPEncrypt creates a PGPEncrypt engine.
func NewPGPEncrypt(g *libkb.GlobalContext, arg *PGPEncryptArg) *PGPEncrypt {
	return &PGPEncrypt{
		arg:          arg,
		Contextified: libkb.NewContextified(g),
	}
}

// Name is the unique engine name.
func (e *PGPEncrypt) Name() string {
	return "PGPEncrypt"
}

// GetPrereqs returns the engine prereqs.
func (e *PGPEncrypt) Prereqs() Prereqs {
	return Prereqs{}
}

// RequiredUIs returns the required UIs.
func (e *PGPEncrypt) RequiredUIs() []libkb.UIKind {
	// context.SecretKeyPromptArg requires SecretUI
	return []libkb.UIKind{
		libkb.SecretUIKind,
		libkb.PgpUIKind,
	}
}

// SubConsumers returns the other UI consumers for this engine.
func (e *PGPEncrypt) SubConsumers() []libkb.UIConsumer {
	return []libkb.UIConsumer{
		&PGPKeyfinder{},
		&ResolveThenIdentify2{},
	}
}

// Run starts the engine.
func (e *PGPEncrypt) Run(m libkb.MetaContext) error {
	// verify valid options based on logged in state:
	ok, uid := isLoggedIn(m)

	if !ok {
		// not logged in.  this is fine, unless they requested signing the message.
		if !e.arg.NoSign {
			return libkb.LoginRequiredError{Context: "you must be logged in to sign"}
		}

		// or trying to encrypt for self
		if !e.arg.NoSelf {
			return libkb.LoginRequiredError{Context: "you must be logged in to encrypt for yourself (or use --no-self flag)"}
		}
	} else {
		me, err := libkb.LoadMeByMetaContextAndUID(m, uid)
		if err != nil {
			return err
		}
		e.me = me
	}

	var mykey *libkb.PGPKeyBundle
	var signer *libkb.PGPKeyBundle
	if !e.arg.NoSign {
		ska := libkb.SecretKeyArg{
			Me:       e.me,
			KeyType:  libkb.PGPKeyType,
			KeyQuery: e.arg.KeyQuery,
		}
		key, err := e.G().Keyrings.GetSecretKeyWithPrompt(m, m.SecretKeyPromptArg(ska, "command-line signature"))
		if err != nil {
			return err
		}

		var ok bool
		mykey, ok = key.(*libkb.PGPKeyBundle)
		if !ok {
			return errors.New("Can only sign with PGP keys")
		}
		signer = mykey
	}

	usernames, err := e.verifyUsers(m, e.arg.Recips, ok)
	if err != nil {
		return err
	}

	kfarg := &PGPKeyfinderArg{
		Usernames: usernames,
	}

	kf := NewPGPKeyfinder(e.G(), kfarg)
	if err := RunEngine2(m, kf); err != nil {
		return err
	}
	uplus := kf.UsersPlusKeys()

	var writer io.WriteCloser
	if e.arg.BinaryOutput {
		writer = e.arg.Sink
	} else {
		aw, err := armor.Encode(e.arg.Sink, "PGP MESSAGE", libkb.PGPArmorHeaders)
		if err != nil {
			return err
		}
		writer = aw
	}

	ks := newKeyset()
	e.warnings = libkb.HashSecurityWarnings{}

	if mykey != nil {
		if w := mykey.SecurityWarnings(
			libkb.HashSecurityWarningOurIdentityHash,
		); len(w) > 0 {
			e.warnings = append(e.warnings, w...)
		}
	}

	for _, up := range uplus {
		for _, k := range up.Keys {
			if len(k.Entity.Revocations)+len(k.Entity.UnverifiedRevocations) > 0 {
				continue
			}

			if w := k.SecurityWarnings(
				libkb.HashSecurityWarningRecipientsIdentityHash,
			); len(w) > 0 {
				e.warnings = append(e.warnings, w...)
			}

			ks.Add(k)
		}
	}

	if len(e.arg.Recips) > 0 && len(ks.keys) == 0 {
		return errors.New("Cannot encrypt - recipient does not have a non-revoked key.")
	}

	if !e.arg.NoSelf {
		if mykey == nil {
			// need to load the public key for the logged in user
			mykey, err = e.loadSelfKey()
			if err != nil {
				return err
			}
		}

		// mykey could still be nil
		if mykey != nil {
			ks.Add(mykey)
		}
	}

	for _, warning := range e.warnings.Strings() {
		if err := m.UIs().PgpUI.OutputPGPWarning(m.Ctx(), keybase1.OutputPGPWarningArg{
			Warning: warning,
		}); err != nil {
			return err
		}
	}

	recipients := ks.Sorted()
	if err := libkb.PGPEncrypt(e.arg.Source, writer, signer, recipients); err != nil {
		return err
	}
	if !e.arg.BinaryOutput {
		return e.arg.Sink.Close()
	}
	return nil
}

func (e *PGPEncrypt) loadSelfKey() (*libkb.PGPKeyBundle, error) {
	me, err := libkb.LoadMe(libkb.NewLoadUserArg(e.G()))
	if err != nil {
		return nil, err
	}

	keys := me.FilterActivePGPKeys(true, e.arg.KeyQuery)
	if len(keys) == 0 {
		return nil, libkb.NoKeyError{Msg: "No PGP key found for encrypting for self (add a PGP key or use --no-self flag)"}
	}
	return keys[0], nil
}

func (e *PGPEncrypt) verifyUsers(m libkb.MetaContext, assertions []string, loggedIn bool) ([]string, error) {
	var names []string
	for _, userAssert := range assertions {
		arg := keybase1.Identify2Arg{
			UserAssertion: userAssert,
			Reason: keybase1.IdentifyReason{
				Type: keybase1.IdentifyReasonType_ENCRYPT,
			},
			AlwaysBlock:      true,
			IdentifyBehavior: keybase1.TLFIdentifyBehavior_CLI,
		}
		eng := NewResolveThenIdentify2(e.G(), &arg)
		if err := RunEngine2(m, eng); err != nil {
			return nil, libkb.IdentifyFailedError{Assertion: userAssert, Reason: err.Error()}
		}
		res, err := eng.Result(m)
		if err != nil {
			return nil, err
		}
		names = append(names, res.Upk.GetName())
	}
	return names, nil
}

// keyset maintains a set of pgp keys, preserving insertion order.
type keyset struct {
	index []keybase1.KID
	keys  map[keybase1.KID]*libkb.PGPKeyBundle
}

// newKeyset creates an empty keyset.
func newKeyset() *keyset {
	return &keyset{keys: make(map[keybase1.KID]*libkb.PGPKeyBundle)}
}

// Add adds bundle to the keyset.  If a key already exists, it
// will be ignored.
func (k *keyset) Add(bundle *libkb.PGPKeyBundle) {
	kid := bundle.GetKID()
	if _, ok := k.keys[kid]; ok {
		return
	}
	k.keys[kid] = bundle
	k.index = append(k.index, kid)
}

// Sorted returns the unique keys in insertion order.
func (k *keyset) Sorted() []*libkb.PGPKeyBundle {
	var sorted []*libkb.PGPKeyBundle
	for _, kid := range k.index {
		sorted = append(sorted, k.keys[kid])
	}
	return sorted
}
