package git

import (
	"testing"

	"github.com/adamwalz/keybase-client/go/externals"
	"github.com/adamwalz/keybase-client/go/kbtest"
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/teams"
)

// Copied from the teams tests.
func SetupTest(tb testing.TB, name string, depth int) (tc libkb.TestContext) {
	tc = libkb.SetupTest(tb, name, depth+1)
	InstallInsecureTriplesec(tc.G)
	tc.G.SetProofServices(externals.NewProofServices(tc.G))
	tc.G.ChatHelper = kbtest.NewMockChatHelper()
	teams.ServiceInit(tc.G)
	return tc
}
