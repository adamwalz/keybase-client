package teams

import (
	"testing"

	"github.com/adamwalz/keybase-client/go/kbtest"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestSetTarsDisabled(t *testing.T) {
	tc := SetupTest(t, "team", 1)
	defer tc.Cleanup()

	_, err := kbtest.CreateAndSignupFakeUser("team", tc.G)
	require.NoError(t, err)

	notifications := kbtest.NewTeamNotifyListener()
	tc.G.SetService()
	tc.G.NotifyRouter.AddListener(notifications)

	name, id := createTeam2(tc)
	t.Logf("Created team %q", name)

	disabled, err := GetTarsDisabled(context.Background(), tc.G, id)
	require.NoError(t, err)
	require.False(t, disabled)

	err = SetTarsDisabled(context.Background(), tc.G, id, true)
	require.NoError(t, err)
	kbtest.CheckTeamMiscNotifications(tc, notifications)

	disabled, err = GetTarsDisabled(context.Background(), tc.G, id)
	require.NoError(t, err)
	require.True(t, disabled)
}
