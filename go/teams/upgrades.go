package teams

import (
	"github.com/adamwalz/keybase-client/go/libkb"
)

type Upgrader struct{}

func NewUpgrader() *Upgrader {
	return &Upgrader{}
}

func (u *Upgrader) Run(m libkb.MetaContext) {
	go BackgroundPinTLFLoop(m)
}
