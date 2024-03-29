package chat

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/adamwalz/keybase-client/go/chat/storage"
	"github.com/adamwalz/keybase-client/go/chat/types"
	"github.com/adamwalz/keybase-client/go/gregor"
	"github.com/adamwalz/keybase-client/go/kbtest"
	"github.com/adamwalz/keybase-client/go/protocol/chat1"
	"github.com/adamwalz/keybase-client/go/protocol/gregor1"
	"github.com/adamwalz/keybase-client/go/protocol/keybase1"
	"github.com/keybase/go-codec/codec"
	"github.com/stretchr/testify/require"
)

func randBytes(t *testing.T, n int) []byte {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	return buf
}

func sendSimple(ctx context.Context, t *testing.T, tc *kbtest.ChatTestContext, ph *PushHandler,
	sender types.Sender, conv chat1.Conversation, user *kbtest.FakeUser,
	iboxXform func(chat1.InboxVers) chat1.InboxVers) {
	uid := gregor1.UID(user.User.GetUID().ToBytes())
	convID := conv.GetConvID()
	outboxID := chat1.OutboxID(randBytes(t, 8))
	nr := tc.G.NotifyRouter
	tc.G.NotifyRouter = nil
	pt := chat1.MessagePlaintext{
		ClientHeader: chat1.MessageClientHeader{
			Conv:        conv.Metadata.IdTriple,
			Sender:      uid,
			TlfName:     user.Username,
			TlfPublic:   false,
			MessageType: chat1.MessageType_TEXT,
			OutboxID:    &outboxID,
		},
		MessageBody: chat1.NewMessageBodyWithText(chat1.MessageText{
			Body: "hi",
		}),
	}
	_, boxed, err := sender.Send(ctx, convID, pt, 0, nil, nil, nil)
	require.NoError(t, err)

	ibox := storage.NewInbox(tc.Context())
	vers, err := ibox.Version(ctx, uid)
	if err != nil {
		require.IsType(t, storage.MissError{}, err)
		vers = 0
	}
	newVers := iboxXform(vers)
	t.Logf("newVers: %d vers: %d", newVers, vers)
	nm := chat1.NewMessagePayload{
		Action:    types.ActionNewMessage,
		ConvID:    conv.GetConvID(),
		Message:   *boxed,
		InboxVers: iboxXform(vers),
		TopicType: chat1.TopicType_CHAT,
	}
	var data []byte
	enc := codec.NewEncoderBytes(&data, &codec.MsgpackHandle{WriteExt: true})
	require.NoError(t, enc.Encode(nm))
	m := gregor1.OutOfBandMessage{
		Uid_:    uid,
		System_: "chat.activity",
		Body_:   data,
	}

	tc.G.NotifyRouter = nr
	require.NoError(t, ph.Activity(ctx, m))
}

func TestPushOrdering(t *testing.T) {
	ctx, world, ri2, _, sender, list := setupTest(t, 1)
	defer world.Cleanup()

	ri := ri2.(*kbtest.ChatRemoteMock)
	u := world.GetUsers()[0]
	uid := u.User.GetUID().ToBytes()
	tc := world.Tcs[u.Username]
	handler := NewPushHandler(tc.Context())
	handler.Start(context.TODO(), nil)
	defer func() { <-handler.Stop(context.TODO()) }()
	handler.SetClock(world.Fc)
	timeout := 2 * time.Second

	conv := newBlankConv(ctx, t, tc, uid, ri, sender, u.Username)
	sendSimple(ctx, t, tc, handler, sender, conv, u,
		func(vers chat1.InboxVers) chat1.InboxVers { return vers + 1 })

	select {
	case <-list.incomingRemote:
	case <-time.After(timeout):
		require.Fail(t, "no notification received")
	}

	sendSimple(ctx, t, tc, handler, sender, conv, u,
		func(vers chat1.InboxVers) chat1.InboxVers { return vers + 2 })
	select {
	case <-list.incomingRemote:
		require.Fail(t, "should not have gotten one of these")
	default:
	}

	sendSimple(ctx, t, tc, handler, sender, conv, u,
		func(vers chat1.InboxVers) chat1.InboxVers { return vers + 1 })
	select {
	case <-list.incomingRemote:
	case <-time.After(timeout):
		require.Fail(t, "no notification received")
	}
	select {
	case <-list.incomingRemote:
	case <-time.After(timeout):
		require.Fail(t, "no notification received")
	}
	handler.orderer.Lock()
	require.Zero(t, len(handler.orderer.waiters))
	handler.orderer.Unlock()

	sendSimple(ctx, t, tc, handler, sender, conv, u,
		func(vers chat1.InboxVers) chat1.InboxVers { return vers + 2 })
	select {
	case <-list.incomingRemote:
		require.Fail(t, "should not have gotten one of these")
	default:
	}

	t.Logf("advancing clock")
	world.Fc.Advance(time.Second)
	select {
	case <-list.incomingRemote:
		require.Fail(t, "not notification expected")
	default:
	}
	world.Fc.Advance(time.Second)
	select {
	case <-list.incomingRemote:
	case <-time.After(timeout):
		require.Fail(t, "no notification received")
	}
	handler.orderer.Lock()
	require.Zero(t, len(handler.orderer.waiters))
	handler.orderer.Unlock()
}

func TestPushAppState(t *testing.T) {
	ctx, world, ri2, _, sender, list := setupTest(t, 1)
	defer world.Cleanup()

	ri := ri2.(*kbtest.ChatRemoteMock)
	u := world.GetUsers()[0]
	uid := u.User.GetUID().ToBytes()
	tc := world.Tcs[u.Username]
	handler := NewPushHandler(tc.Context())
	handler.Start(context.TODO(), nil)
	defer func() { <-handler.Stop(context.TODO()) }()
	handler.SetClock(world.Fc)
	conv := newBlankConv(ctx, t, tc, uid, ri, sender, u.Username)

	tc.G.MobileAppState.Update(keybase1.MobileAppState_BACKGROUND)
	sendSimple(ctx, t, tc, handler, sender, conv, u,
		func(vers chat1.InboxVers) chat1.InboxVers { return vers + 1 })
	select {
	case <-list.incomingRemote:
	case <-time.After(20 * time.Second):
		require.Fail(t, "no message received")
	}
	tc.G.MobileAppState.Update(keybase1.MobileAppState_FOREGROUND)
	sendSimple(ctx, t, tc, handler, sender, conv, u,
		func(vers chat1.InboxVers) chat1.InboxVers { return vers + 1 })
	select {
	case <-list.incomingRemote:
	case <-time.After(20 * time.Second):
		require.Fail(t, "no message received")
	}
}

func makeTypingNotification(t *testing.T, uid gregor1.UID, convID chat1.ConversationID, typing bool) gregor.OutOfBandMessage {

	nm := chat1.RemoteUserTypingUpdate{
		Uid:    uid,
		ConvID: convID,
		Typing: typing,
	}
	var data []byte
	enc := codec.NewEncoderBytes(&data, &codec.MsgpackHandle{WriteExt: true})
	require.NoError(t, enc.Encode(nm))
	m := gregor1.OutOfBandMessage{
		Uid_:    uid,
		System_: "chat.typing",
		Body_:   data,
	}
	return m
}

func TestPushTyping(t *testing.T) {
	ctx, world, ri2, _, sender, list := setupTest(t, 1)
	defer world.Cleanup()

	ri := ri2.(*kbtest.ChatRemoteMock)
	u := world.GetUsers()[0]
	uid := u.User.GetUID().ToBytes()
	tc := world.Tcs[u.Username]
	handler := NewPushHandler(tc.Context())
	handler.Start(context.TODO(), nil)
	defer func() { <-handler.Stop(context.TODO()) }()
	handler.SetClock(world.Fc)
	handler.typingMonitor.SetClock(world.Fc)
	handler.typingMonitor.SetTimeout(time.Minute)

	conv := newBlankConv(ctx, t, tc, uid, ri, sender, u.Username)

	confirmTyping := func(list *chatListener) {
		select {
		case updates := <-list.typingUpdate:
			require.Equal(t, 1, len(updates))
			require.Equal(t, conv.GetConvID(), updates[0].ConvID)
			require.Equal(t, 1, len(updates[0].Typers))
			require.Equal(t, uid, updates[0].Typers[0].Uid.ToBytes())
		case <-time.After(20 * time.Second):
			require.Fail(t, "no typing notification")
		}
	}

	confirmNotTyping := func(list *chatListener) {
		select {
		case updates := <-list.typingUpdate:
			require.Equal(t, 1, len(updates))
			require.Equal(t, conv.GetConvID(), updates[0].ConvID)
			require.Zero(t, len(updates[0].Typers))
		case <-time.After(2 * time.Second):
			require.Fail(t, "no typing notification")
		}
	}

	t.Logf("test basic")
	err := handler.Typing(context.TODO(), makeTypingNotification(t, uid, conv.GetConvID(), true))
	require.NoError(t, err)
	confirmTyping(list)
	err = handler.Typing(context.TODO(), makeTypingNotification(t, uid, conv.GetConvID(), false))
	require.NoError(t, err)
	confirmNotTyping(list)

	t.Logf("test expiration")
	err = handler.Typing(context.TODO(), makeTypingNotification(t, uid, conv.GetConvID(), true))
	require.NoError(t, err)
	confirmTyping(list)
	world.Fc.Advance(time.Hour)
	confirmNotTyping(list)

	t.Logf("test extend")
	extendCh := make(chan struct{})
	handler.typingMonitor.extendCh = &extendCh
	err = handler.Typing(context.TODO(), makeTypingNotification(t, uid, conv.GetConvID(), true))
	require.NoError(t, err)
	confirmTyping(list)
	world.Fc.Advance(30 * time.Second)
	err = handler.Typing(context.TODO(), makeTypingNotification(t, uid, conv.GetConvID(), true))
	require.NoError(t, err)
	select {
	case <-list.typingUpdate:
		require.Fail(t, "should have extended")
	default:
	}
	select {
	case <-extendCh:
	case <-time.After(20 * time.Second):
		require.Fail(t, "no extend callback")
	}
	world.Fc.Advance(40 * time.Second)
	select {
	case <-list.typingUpdate:
		require.Fail(t, "not far enough")
	default:
	}
	world.Fc.Advance(40 * time.Second)
	confirmNotTyping(list)

	require.Zero(t, len(handler.typingMonitor.typers))
}
