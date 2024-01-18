package teams

import (
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/teams/hidden"
)

func ServiceInit(g *libkb.GlobalContext) {
	NewTeamLoaderAndInstall(g)
	NewFastTeamLoaderAndInstall(g)
	NewAuditorAndInstall(g)
	NewBoxAuditorAndInstall(g)
	NewImplicitTeamConflictInfoCacheAndInstall(g)
	NewImplicitTeamCacheAndInstall(g)
	hidden.NewChainManagerAndInstall(g)
	NewTeamRoleMapManagerAndInstall(g)
}
