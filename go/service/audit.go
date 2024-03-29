package service

import (
	"github.com/adamwalz/keybase-client/go/libkb"
	keybase1 "github.com/adamwalz/keybase-client/go/protocol/keybase1"
	"github.com/adamwalz/keybase-client/go/teams"
	"github.com/keybase/go-framed-msgpack-rpc/rpc"

	"golang.org/x/net/context"
)

type AuditHandler struct {
	libkb.Contextified
	*BaseHandler
}

func NewAuditHandler(xp rpc.Transporter, g *libkb.GlobalContext) *AuditHandler {
	handler := &AuditHandler{
		Contextified: libkb.NewContextified(g),
		BaseHandler:  NewBaseHandler(g, xp),
	}
	return handler
}

var _ keybase1.AuditInterface = (*AuditHandler)(nil)

func (h *AuditHandler) IsInJail(ctx context.Context, arg keybase1.IsInJailArg) (ret bool, err error) {
	mctx := libkb.NewMetaContext(ctx, h.G())
	defer mctx.Trace("AuditHandler#IsInJail", &err)()
	return h.G().GetTeamBoxAuditor().IsInJail(mctx, arg.TeamID)
}

func (h *AuditHandler) BoxAuditTeam(ctx context.Context, arg keybase1.BoxAuditTeamArg) (res *keybase1.BoxAuditAttempt, err error) {
	mctx := libkb.NewMetaContext(ctx, h.G())
	defer mctx.Trace("AuditHandler#BoxAuditTeam", &err)()
	return h.G().GetTeamBoxAuditor().BoxAuditTeam(mctx, arg.TeamID)
}

func (h *AuditHandler) AttemptBoxAudit(ctx context.Context, arg keybase1.AttemptBoxAuditArg) (res keybase1.BoxAuditAttempt, err error) {
	mctx := libkb.NewMetaContext(ctx, h.G())
	defer mctx.Trace("AuditHandler#AttemptBoxAudit", &err)()
	return h.G().GetTeamBoxAuditor().Attempt(mctx, arg.TeamID, arg.RotateBeforeAudit), nil
}

func (h *AuditHandler) KnownTeamIDs(ctx context.Context, sessionID int) (res []keybase1.TeamID, err error) {
	mctx := libkb.NewMetaContext(ctx, h.G())
	defer mctx.Trace("AuditHandler#KnownTeamIDs", &err)()
	return teams.KnownTeamIDs(mctx)
}
