package client

import (
	"fmt"
	"strings"

	"github.com/keybase/cli"
	"github.com/adamwalz/keybase-client/go/libcmdline"
	"github.com/adamwalz/keybase-client/go/libkb"
	"github.com/adamwalz/keybase-client/go/protocol/chat1"
	keybase1 "github.com/adamwalz/keybase-client/go/protocol/keybase1"
	"golang.org/x/net/context"
)

type CmdTeamSettings struct {
	libkb.Contextified

	Team   keybase1.TeamName
	teamID keybase1.TeamID

	// These fields are non-zero valued when their action is requested
	Description           *string
	WelcomeMessage        *string
	ResetWelcomeMessage   *bool
	JoinAsRole            *keybase1.TeamRole
	ProfilePromote        *bool
	AllowProfilePromote   *bool
	Showcase              *bool
	DisableAccessRequests *bool
}

func newCmdTeamSettings(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	return cli.Command{
		Name:         "settings",
		ArgumentHelp: "<team name>",
		Usage:        "Edit team settings.",
		Examples: `
Review team settings:
    keybase team settings acme
Open a team so anyone can join as a reader:
    keybase team settings acme --open-team=reader
Showcase a team publicly:
    keybase team settings acme --showcase=yes
Promote a team on your profile:
    keybase team settings acme --profile-promote=yes
Set a description for the team to show if promoted:
    keybase team settings acme --description="Rocket-Powered Products"
Clear the team description:
    keybase team settings acme --description=""
`,
		Action: func(c *cli.Context) {
			cmd := NewCmdTeamSettingsRunner(g)
			cl.ChooseCommand(cmd, "settings", c)
		},
		Flags: []cli.Flag{
			// Many of these are StringFlag instead of BoolFlag because BoolFlag is displeasing.
			// For example `keybase team settings teamname --bool-flag false` sets the flag to true.

			cli.BoolFlag{
				Name:  "p, print",
				Usage: "Print all your team settings",
			},
			cli.StringFlag{
				Name:  "description",
				Usage: "Set the team description",
			},
			cli.StringFlag{
				Name:  "open-team",
				Usage: "[reader|writer|off] Set whether anyone can join without being invited and the role they become",
			},
			cli.StringFlag{
				Name:  "profile-promote",
				Usage: "[yes|no] Set whether your own profile should promote this team and its description",
			},
			cli.StringFlag{
				Name:  "allow-profile-promote",
				Usage: "[yes|no] Set whether non-admins are allowed to promote this team and its description on their profiles",
			},
			cli.StringFlag{
				Name:  "showcase",
				Usage: "[yes|no] Set whether to promote this team and its description on keybase.io/popular-teams",
			},
			cli.StringFlag{
				Name:  "disable-access-requests",
				Usage: "[yes|no] Set whether it should be possible to access request to this team",
			},
			// cli.StringFlag{
			// 	Name:  "welcome-message",
			// 	Usage: "Set a welcome message for new team members. Empty string for no welcome message.",
			// },
			// cli.BoolFlag{
			// 	Name:  "reset-welcome-message",
			// 	Usage: "Reset the welcome message to the default.",
			// },
		},
	}
}

func NewCmdTeamSettingsRunner(g *libkb.GlobalContext) *CmdTeamSettings {
	return &CmdTeamSettings{Contextified: libkb.NewContextified(g)}
}

func (c *CmdTeamSettings) ParseArgv(ctx *cli.Context) (err error) {
	c.Team, err = ParseOneTeamNameK1(ctx)
	if err != nil {
		return err
	}

	var exclusiveActions []string

	if ctx.IsSet("description") {
		exclusiveActions = append(exclusiveActions, "description")
		desc := ctx.String("description")
		c.Description = &desc
	}

	if ctx.IsSet("open-team") {
		exclusiveActions = append(exclusiveActions, "open-team")

		role := keybase1.TeamRole_NONE
		val := ctx.String("open-team")
		switch val {
		case "reader":
			role = keybase1.TeamRole_READER
		case "writer":
			role = keybase1.TeamRole_WRITER
		default:
			open, err := cli.ParseBoolFriendly(val)
			if err != nil || open {
				return fmt.Errorf("open-team must be one of [reader|writer|off]")
			}
		}
		c.JoinAsRole = &role
	}

	if ctx.IsSet("profile-promote") {
		exclusiveActions = append(exclusiveActions, "profile-promote")
		val, err := ctx.BoolStrict("profile-promote")
		if err != nil {
			return err
		}
		c.ProfilePromote = &val
	}

	if ctx.IsSet("allow-profile-promote") {
		exclusiveActions = append(exclusiveActions, "allow-profile-promote")
		val, err := ctx.BoolStrict("allow-profile-promote")
		if err != nil {
			return err
		}
		c.AllowProfilePromote = &val
	}

	if ctx.IsSet("showcase") {
		exclusiveActions = append(exclusiveActions, "showcase")
		val, err := ctx.BoolStrict("showcase")
		if err != nil {
			return err
		}
		c.Showcase = &val
	}

	if ctx.IsSet("disable-access-requests") {
		exclusiveActions = append(exclusiveActions, "disable-access-requests")
		val, err := ctx.BoolStrict("disable-access-requests")
		if err != nil {
			return err
		}
		c.DisableAccessRequests = &val
	}

	if ctx.IsSet("welcome-message") {
		exclusiveActions = append(exclusiveActions, "welcome-message")
		welcomeMessage := ctx.String("welcome-message")
		c.WelcomeMessage = &welcomeMessage
	}

	if ctx.IsSet("reset-welcome-message") {
		exclusiveActions = append(exclusiveActions, "reset-welcome-message")
		resetWelcomeMessage := ctx.Bool("reset-welcome-message")
		if !resetWelcomeMessage {
			return fmt.Errorf("cannot pass a false value to --reset-welcome-message")
		}
		c.ResetWelcomeMessage = &resetWelcomeMessage
	}

	if len(exclusiveActions) > 1 {
		return fmt.Errorf("only one of these actions a time: %v", strings.Join(exclusiveActions, ", "))
	}

	return nil
}

func (c *CmdTeamSettings) Run() error {
	ctx, ctxCancel := context.WithCancel(context.TODO())
	defer ctxCancel()
	ctx = libkb.WithLogTag(ctx, "CTS")

	cli, err := GetTeamsClient(c.G())
	if err != nil {
		return err
	}

	teamID, err := cli.GetTeamID(context.Background(), c.Team.String())
	if err != nil {
		return err
	}
	c.teamID = teamID

	if c.Description != nil {
		err = c.setDescription(ctx, cli, *c.Description)
		if err != nil {
			return err
		}
	}

	if c.JoinAsRole != nil {
		err = c.setOpenness(ctx, cli, *c.JoinAsRole)
		if err != nil {
			return err
		}
	}

	if c.ProfilePromote != nil {
		err = c.setProfilePromote(ctx, cli, *c.ProfilePromote)
		if err != nil {
			return err
		}
	}

	if c.AllowProfilePromote != nil {
		err = c.setAllowProfilePromote(ctx, cli, *c.AllowProfilePromote)
		if err != nil {
			return err
		}
	}

	if c.Showcase != nil {
		err = c.setShowcase(ctx, cli, *c.Showcase)
		if err != nil {
			return err
		}
	}

	if c.DisableAccessRequests != nil {
		err = c.setDisableAccessRequests(ctx, cli, *c.DisableAccessRequests)
		if err != nil {
			return err
		}
	}

	if c.WelcomeMessage != nil {
		err = c.setWelcomeMessage(ctx, *c.WelcomeMessage)
		if err != nil {
			return err
		}
	}

	if c.ResetWelcomeMessage != nil {
		err = c.resetWelcomeMessage(ctx)
		if err != nil {
			return err
		}
	}

	err = c.printCurrentSettings(ctx, cli)
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTeamSettings) setDescription(ctx context.Context, cli keybase1.TeamsClient, desc string) error {
	return cli.SetTeamShowcase(ctx, keybase1.SetTeamShowcaseArg{
		TeamID:            c.teamID,
		IsShowcased:       nil,
		Description:       &desc,
		AnyMemberShowcase: nil,
	})
}

func (c *CmdTeamSettings) setOpenness(ctx context.Context, cli keybase1.TeamsClient, joinAsRole keybase1.TeamRole) error {
	dui := c.G().UI.GetTerminalUI()

	open := joinAsRole != keybase1.TeamRole_NONE

	arg := keybase1.TeamSetSettingsArg{
		TeamID: c.teamID,
		Settings: keybase1.TeamSettings{
			Open:   open,
			JoinAs: joinAsRole,
		},
	}

	err := cli.TeamSetSettings(ctx, arg)
	if err != nil {
		if e, ok := err.(libkb.NoOpError); ok {
			dui.Printf("%s\n", e.Desc)
			return nil
		}

		return err
	}

	if open {
		dui.Printf("Team set to open.\n")
	} else {
		dui.Printf("Team set to closed.\n")
	}
	return nil
}

func (c *CmdTeamSettings) setProfilePromote(ctx context.Context, cli keybase1.TeamsClient, promote bool) error {
	err := cli.SetTeamMemberShowcase(ctx, keybase1.SetTeamMemberShowcaseArg{
		TeamID:      c.teamID,
		IsShowcased: promote,
	})
	if err != nil {
		if _, ok := err.(libkb.TeamBadMembershipError); ok {
			dui := c.G().UI.GetTerminalUI()
			dui.Printf("You cannot promote team %q on your profile because you\n", c.Team)
			dui.Printf("are not a member of the team.\n\n")
			dui.Printf("You can add yourself to the team with `keybase team add-member ...`\n")
			dui.Printf("and then you can promote the team on your profile.\n\n")
			return nil
		}
		return err
	}

	return nil
}

func (c *CmdTeamSettings) setAllowProfilePromote(ctx context.Context, cli keybase1.TeamsClient, allow bool) error {
	return cli.SetTeamShowcase(ctx, keybase1.SetTeamShowcaseArg{
		TeamID:            c.teamID,
		IsShowcased:       nil,
		Description:       nil,
		AnyMemberShowcase: &allow,
	})
}

func (c *CmdTeamSettings) setShowcase(ctx context.Context, cli keybase1.TeamsClient, showcase bool) error {
	return cli.SetTeamShowcase(ctx, keybase1.SetTeamShowcaseArg{
		TeamID:            c.teamID,
		IsShowcased:       &showcase,
		Description:       nil,
		AnyMemberShowcase: nil,
	})
}

func (c *CmdTeamSettings) setDisableAccessRequests(ctx context.Context, cli keybase1.TeamsClient, disabled bool) error {
	return cli.SetTarsDisabled(ctx, keybase1.SetTarsDisabledArg{
		TeamID:   c.teamID,
		Disabled: disabled,
	})
}

func (c *CmdTeamSettings) setWelcomeMessage(ctx context.Context, welcomeMessage string) error {
	if err := CheckAndStartStandaloneChat(c.G(), chat1.ConversationMembersType_TEAM); err != nil {
		return err
	}
	msg := chat1.WelcomeMessage{Set: true, Raw: welcomeMessage}
	cli, err := GetChatLocalClient(c.G())
	if err != nil {
		return err
	}
	return cli.SetWelcomeMessage(ctx, chat1.SetWelcomeMessageArg{
		TeamID:  c.teamID,
		Message: msg,
	})
}

func (c *CmdTeamSettings) resetWelcomeMessage(ctx context.Context) error {
	if err := CheckAndStartStandaloneChat(c.G(), chat1.ConversationMembersType_TEAM); err != nil {
		return err
	}
	cli, err := GetChatLocalClient(c.G())
	if err != nil {
		return err
	}
	return cli.SetWelcomeMessage(ctx, chat1.SetWelcomeMessageArg{
		TeamID:  c.teamID,
		Message: chat1.WelcomeMessage{Set: false},
	})
}

func (c *CmdTeamSettings) printCurrentSettings(ctx context.Context, cli keybase1.TeamsClient) error {
	res, err := cli.GetAnnotatedTeamByName(ctx, c.Team.String())
	if err != nil {
		return err
	}

	var showcaseInfo *keybase1.TeamAndMemberShowcase
	tmp, err := cli.GetTeamAndMemberShowcase(ctx, c.teamID)
	if err != nil {
		c.G().Log.CDebugf(ctx, "failed to get team showcase info: %v", err)
	} else {
		showcaseInfo = &tmp
	}

	dui := c.G().UI.GetTerminalUI()
	dui.Printf("Current settings for team %q:\n", c.Team.String())
	if showcaseInfo != nil && showcaseInfo.TeamShowcase.Description != nil {
		dui.Printf("  Description:             %v\n", *showcaseInfo.TeamShowcase.Description)
	}
	dui.Printf("  Open:                     %v\n", c.tfToYn(res.Settings.Open,
		fmt.Sprintf("default membership = %v", strings.ToLower(res.Settings.JoinAs.String()))))
	if showcaseInfo != nil {
		dui.Printf("  Showcased:                %v\n", c.tfToYn(showcaseInfo.TeamShowcase.IsShowcased, "on keybase.io/popular-teams"))
		dui.Printf("  Promoted:                 %v\n", c.tfToYn(showcaseInfo.IsMemberShowcased, "on your profile"))
		dui.Printf("  Non-admins can promote:   %v\n", c.tfToYn(showcaseInfo.TeamShowcase.AnyMemberShowcase, "on their profiles"))
	}

	// TarsDisabled info is only available for owners and admins, check
	// if we can make this call so we don't make a GetTarsDisabled call
	// that results in an error which is sent to the console.
	ret, err := cli.CanUserPerform(ctx, c.Team.String())
	if err != nil {
		c.G().Log.CDebugf(ctx, "failed to get CanUserPerform info: %v", err)
	} else {
		if ret.ChangeTarsDisabled {
			ok, err := cli.GetTarsDisabled(ctx, c.teamID)
			if err != nil {
				c.G().Log.CDebugf(ctx, "failed to call GetTarsEnabled: %v", err)
			} else {
				dui.Printf("  Access requests disabled: %v\n", c.tfToYn(ok, ""))
			}
		}
	}

	err = CheckAndStartStandaloneChat(c.G(), chat1.ConversationMembersType_TEAM)
	if err != nil {
		dui.Printf("  Welcome message: [failed to start chat system, not available in standalone mode]\n")
	} else {
		chatCli, err := GetChatLocalClient(c.G())
		if err != nil {
			return err
		}
		msg, err := chatCli.GetWelcomeMessage(ctx, c.teamID)
		if err != nil {
			c.G().Log.CWarningf(ctx, "failed to call get welcome message: %v", err)
		} else {
			if msg.Set {
				if len(msg.Raw) > 0 {
					dui.Printf("  Welcome message:          %q\n", msg.Raw)
				} else {
					dui.Printf("  Welcome message:          none\n")
				}
			} else {
				dui.Printf("  Welcome message:          unset (default)\n")

			}
		}
	}

	return nil
}

func (c *CmdTeamSettings) tfToYn(x bool, parenthetical string) string {
	if x {
		if len(parenthetical) > 0 {
			return fmt.Sprintf("yes (%v)", parenthetical)
		}
		return "yes"
	}
	return "no"
}

func (c *CmdTeamSettings) GetUsage() libkb.Usage {
	return libkb.Usage{
		Config:    true,
		API:       true,
		KbKeyring: true,
	}
}
