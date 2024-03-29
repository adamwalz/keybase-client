// Copyright 2017 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"context"
	"errors"
	"github.com/keybase/cli"
	"github.com/adamwalz/keybase-client/go/libcmdline"
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/protocol/chat1"
	"github.com/adamwalz/keybase-client/go/protocol/keybase1"
)

type CmdTeamAddMember struct {
	libkb.Contextified
	Team                 string
	Email                string
	Phone                string
	Username             string
	Role                 keybase1.TeamRole
	BotSettings          *keybase1.TeamBotSettings
	SkipChatNotification bool
	EmailInviteMessage   *string
}

func newCmdTeamAddMember(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	cmd := cli.Command{
		Name:         "add-member",
		ArgumentHelp: "<team name>",
		Usage:        "Add a user to a team.",
		Action: func(c *cli.Context) {
			cl.ChooseCommand(NewCmdTeamAddMemberRunner(g), "add-member", c)
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "u, user",
				Usage: "username",
			},
			cli.StringFlag{
				Name:  "e, email",
				Usage: "email address to invite",
			},
			cli.StringFlag{
				Name:  "p, phone",
				Usage: "phone number to invite",
			},
			cli.StringFlag{
				Name:  "r, role",
				Usage: "team role (owner, admin, writer, reader, bot, restrictedbot) [required]",
			},
			cli.BoolFlag{
				Name:  "s, skip-chat-message",
				Usage: "skip chat welcome message",
			},
			cli.StringFlag{
				Name:  "m, email-invite-message",
				Usage: "send a welcome message along with your email invitation",
			},
		},
		Description: teamAddMemberDoc,
	}

	cmd.Flags = append(cmd.Flags, botSettingsFlags...)
	return cmd
}

func NewCmdTeamAddMemberRunner(g *libkb.GlobalContext) *CmdTeamAddMember {
	return &CmdTeamAddMember{Contextified: libkb.NewContextified(g)}
}

func (c *CmdTeamAddMember) ParseArgv(ctx *cli.Context) error {
	var err error
	c.Team, err = ParseOneTeamName(ctx)
	if err != nil {
		return err
	}

	c.Role, err = ParseRole(ctx)
	if err != nil {
		return err
	}

	emailInviteMsg := ctx.String("email-invite-message")
	if len(emailInviteMsg) > 0 {
		c.EmailInviteMessage = &emailInviteMsg
	}

	c.Email = ctx.String("email")
	if len(c.Email) > 0 {
		if !libkb.CheckEmail.F(c.Email) {
			return errors.New("invalid email address")
		}
		return nil
	}

	c.Phone = ctx.String("phone")
	if len(c.Phone) > 0 {
		return nil
	}

	c.Username, err = ParseUser(ctx)
	if err != nil {
		return err
	}

	c.SkipChatNotification = ctx.Bool("skip-chat-message")

	if c.Role.IsRestrictedBot() {
		c.BotSettings = ParseBotSettings(ctx)
	}

	return nil
}

func (c *CmdTeamAddMember) Run() error {
	cli, err := GetTeamsClient(c.G())
	if err != nil {
		return err
	}

	if err := ValidateBotSettingsConvs(c.G(), c.Team,
		chat1.ConversationMembersType_TEAM, c.BotSettings); err != nil {
		return err
	}

	teamID, err := cli.GetTeamID(context.Background(), c.Team)
	if err != nil {
		return err
	}

	arg := keybase1.TeamAddMemberArg{
		TeamID:               teamID,
		Email:                c.Email,
		Phone:                c.Phone,
		Username:             c.Username,
		Role:                 c.Role,
		BotSettings:          c.BotSettings,
		SendChatNotification: !c.SkipChatNotification,
		EmailInviteMessage:   c.EmailInviteMessage,
	}

	res, err := cli.TeamAddMember(context.Background(), arg)
	if err != nil {
		return err
	}

	dui := c.G().UI.GetDumbOutputUI()
	if !res.Invited {
		// TeamAddMember resulted in the user added to the team
		if c.Email != "" {
			dui.Printf("%s matched the Keybase username %s.\n", c.Email, res.User.Username)
		} else if c.Phone != "" {
			dui.Printf("%s matched the Keybase username %s.\n", c.Phone, res.User.Username)
		}
		if res.ChatSending {
			// The chat message may still be in flight or fail.
			dui.Printf("Success! A keybase chat message has been sent to %s. To skip this, use `-s` or `--skip-chat-message`\n", res.User.Username)
		} else {
			dui.Printf("Success! %s added to team.\n", res.User.Username)
		}
		return nil
	}

	// TeamAddMember resulted in the user invited to the team

	if c.Email != "" {
		// email invitation
		dui.Printf("Pending! Email sent to %s with signup instructions. When they join you will be notified.\n", c.Email)
		return nil
	}

	if c.Phone != "" {
		// phone invitation
		dui.Printf("Pending! When %s joins Keybase and proves their phone number, they will be added to the team automatically.\n", c.Phone)
		return nil
	}

	if res.User != nil {
		// user without keys or without puk
		dui.Printf("Pending! Keybase stored a team invitation for %s. When they open the Keybase app, their account will be upgraded and you will be notified.\n", res.User.Username)
	} else {
		// "sharing before signup" user
		dui.Printf("Pending! Keybase stored a team invitation for %s. When they join Keybase you will be notified.\n", c.Username)
	}

	return nil
}

func (c *CmdTeamAddMember) GetUsage() libkb.Usage {
	return libkb.Usage{
		Config:    true,
		API:       true,
		KbKeyring: true,
	}
}

const teamAddMemberDoc = `"keybase team add-member" allows you to add users to a team.

EXAMPLES:

Add an existing keybase user:

    keybase team add-member acme --user=alice --role=writer

Add a user via social assertion:

    keybase team add-member acme --user=alice@twitter --role=writer

Add a user via email:

    keybase team add-member acme --email=alice@mail.com --role=reader

Add a user via phone:

    keybase team add-member acme --phone=18581234567 --role=reader
`
