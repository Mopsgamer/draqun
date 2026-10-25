package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestChannelsManagementAndMessaging(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner User and create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Channel Owner", "ch_owner", "ch_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Channel Test Group", "channel_test_group")

	// Verify group has no channels by default
	channels, err := model.GetGroupChannels(group.Id)
	if err != nil {
		t.Fatalf("failed to query group channels: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("expected group to have 0 channels on creation by default, got %d", len(channels))
	}

	// 1. GET /groups/:id/settings/channels
	respChannelsPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/channels", group.Id))
	if err != nil {
		t.Fatalf("GET channels settings page failed: %v", err)
	}
	defer respChannelsPage.Close()

	if respChannelsPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for channels settings page, got %d", respChannelsPage.StatusCode())
	}
	if !strings.Contains(respChannelsPage.String(), "Group Channels") {
		t.Errorf("expected channels page content, got: %s", respChannelsPage.String())
	}

	// 1b. GET /groups/:id/settings/channels/create
	respCreatePage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/channels/create", group.Id))
	if err != nil {
		t.Fatalf("GET channel create page failed: %v", err)
	}
	defer respCreatePage.Close()

	if respCreatePage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for channel create page, got %d", respCreatePage.StatusCode())
	}
	if !strings.Contains(respCreatePage.String(), "Create Channel") {
		t.Errorf("expected channel create page content, got: %s", respCreatePage.String())
	}

	// 2. POST /groups/:id/channels/create -> Create main channel
	respCreateMain, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "general",
			"moniker":     "General Chat",
			"description": "General text chat",
			"type":        "text",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create channel failed: %v", err)
	}
	respCreateMain.Close()

	if respCreateMain.StatusCode() != http.StatusOK && respCreateMain.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200/303 for channel creation, got %d", respCreateMain.StatusCode())
	}

	// Verify channel was created
	channels, err = model.GetGroupChannels(group.Id)
	if err != nil || len(channels) != 1 {
		t.Fatalf("expected 1 group channel, got %d (err: %v)", len(channels), err)
	}
	mainChannel := channels[0]
	if mainChannel.Moniker != "General Chat" {
		t.Errorf("expected moniker 'General Chat', got '%s'", mainChannel.Moniker)
	}

	// 3. POST /groups/:id/channels/create -> Create subchannel
	respCreateSub, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "announcements",
			"moniker":     "Announcements Subchannel",
			"description": "Important announcements",
			"type":        "text",
			"parent_id":   fmt.Sprintf("%d", mainChannel.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create subchannel failed: %v", err)
	}
	respCreateSub.Close()

	subchannels, err := model.GetSubchannels(mainChannel.Id)
	if err != nil || len(subchannels) != 1 {
		t.Fatalf("expected 1 subchannel under mainChannel, got %d (err: %v)", len(subchannels), err)
	}
	subChannel := subchannels[0]
	if subChannel.Moniker != "Announcements Subchannel" {
		t.Errorf("expected subchannel moniker 'Announcements Subchannel', got '%s'", subChannel.Moniker)
	}

	// 3b. Create another top-level channel to test moving/reordering
	respCreateSecond, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "random",
			"moniker":     "Random Chat",
			"description": "Random discussion",
			"type":        "text",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create second channel failed: %v", err)
	}
	respCreateSecond.Close()

	topChannels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(topChannels) != 2 {
		t.Fatalf("expected 2 top-level channels, got %d", len(topChannels))
	}
	if topChannels[0].Id != mainChannel.Id {
		t.Errorf("expected first channel to be mainChannel (%d), got %d", mainChannel.Id, topChannels[0].Id)
	}

	// Move mainChannel down
	respMoveDown, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/channels/%d/move?direction=down", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("move channel down failed: %v", err)
	}
	respMoveDown.Close()

	topChannelsAfterMove, _ := model.GetGroupChannels(group.Id)
	if len(topChannelsAfterMove) == 2 && topChannelsAfterMove[0].Id == mainChannel.Id {
		t.Errorf("expected mainChannel to move down to index 1, but it's still at index 0")
	}

	// Move mainChannel back up
	respMoveUp, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/channels/%d/move?direction=up", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("move channel up failed: %v", err)
	}
	respMoveUp.Close()

	topChannelsBackUp, _ := model.GetGroupChannels(group.Id)
	if len(topChannelsBackUp) == 2 && topChannelsBackUp[0].Id != mainChannel.Id {
		t.Errorf("expected mainChannel to move back up to index 0")
	}

	// 4. POST /groups/:id/channels/:channel_id/messages/create -> Post message to main channel
	respMsg, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"content": "Hello from General Chat!",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("post message to channel failed: %v", err)
	}
	respMsg.Close()

	if respMsg.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for message creation in channel, got %d", respMsg.StatusCode())
	}

	// Verify message in channel
	msgs := mainChannel.MessagesPage(1, 10)
	if len(msgs) != 1 || msgs[0].Content != "Hello from General Chat!" {
		t.Errorf("expected 1 message in channel with content 'Hello from General Chat!', got %v", msgs)
	}

	// 5. GET /groups/:id/channels/:channel_id/messages/page/1 -> Fetch paginated messages
	respGetMsgs, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("GET channel messages page failed: %v", err)
	}
	defer respGetMsgs.Close()

	if respGetMsgs.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for channel messages page, got %d", respGetMsgs.StatusCode())
	}

	// 6. Test Permission Rule Creation
	respSetPerm, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":     "read",
			"level":    "own",
			"priority": "10",
		}).
		Post(fmt.Sprintf("/groups/%d/permissions", group.Id))
	if err != nil {
		t.Fatalf("create permission rule failed: %v", err)
	}
	respSetPerm.Close()

	// 7. GET /chat/groups/:id/channels/:channel_id -> View channel chat page
	respViewChannel, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("GET chat channel page failed: %v", err)
	}
	defer respViewChannel.Close()

	if respViewChannel.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for chat channel page, got %d", respViewChannel.StatusCode())
	}
	if !strings.Contains(respViewChannel.String(), "General Chat") {
		t.Errorf("expected chat channel page to contain channel moniker 'General Chat', got: %s", respViewChannel.String())
	}

	// 8. DELETE /groups/:id/channels/:channel_id -> Delete subchannel
	respDeleteSub, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/channels/%d", group.Id, subChannel.Id))
	if err != nil {
		t.Fatalf("delete subchannel failed: %v", err)
	}
	respDeleteSub.Close()

	subchannelsAfter, _ := model.GetSubchannels(mainChannel.Id)
	if len(subchannelsAfter) != 0 {
		t.Errorf("expected 0 subchannels after deletion, got %d", len(subchannelsAfter))
	}
}

func TestDirectMessagingFlow(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookie1, user1 := createTestUserCookie(t, ts, "DM User One", "dm_user1", "dm_user1@example.com")
	cookie2, user2 := createTestUserCookie(t, ts, "DM User Two", "dm_user2", "dm_user2@example.com")
	cookie3, user3 := createTestUserCookie(t, ts, "DM User Three", "dm_user3", "dm_user3@example.com")

	// 1. Self DM attempt -> should fail
	respSelfDM, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/friends/dm/%d", user1.Id))
	if err != nil {
		t.Fatalf("self DM request failed: %v", err)
	}
	defer respSelfDM.Close()
	if !strings.Contains(respSelfDM.String(), "cannot direct message yourself") {
		t.Errorf("expected self DM prevention alert, got: %s", respSelfDM.String())
	}

	// 2. User 1 initiates DM with User 2
	respCreateDM, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		Post(fmt.Sprintf("/friends/dm/%d", user2.Id))
	if err != nil {
		t.Fatalf("DM creation request failed: %v", err)
	}
	defer respCreateDM.Close()

	if respCreateDM.StatusCode() != http.StatusOK && respCreateDM.StatusCode() != http.StatusSeeOther && respCreateDM.StatusCode() != http.StatusFound {
		t.Fatalf("expected redirect or 200 OK for DM creation, got %d", respCreateDM.StatusCode())
	}

	// Verify direct channel exists in DB
	dmChannel, err := model.GetOrCreateDirectChannel(user1.Id, user2.Id)
	if err != nil || dmChannel == nil || dmChannel.IsEmpty() {
		t.Fatalf("expected direct channel to exist in DB, got: %v (err: %v)", dmChannel, err)
	}

	// 3. GET /chat/dm/:channel_id for participants vs non-participant
	respViewDM1, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		Get(fmt.Sprintf("/chat/dm/%d", dmChannel.Id))
	if err != nil || respViewDM1.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for participant user1 viewing DM page, got status: %v", respViewDM1.StatusCode())
	}
	respViewDM1.Close()

	respViewDM2, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		Get(fmt.Sprintf("/chat/dm/%d", dmChannel.Id))
	if err != nil || respViewDM2.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for participant user2 viewing DM page, got status: %v", respViewDM2.StatusCode())
	}
	respViewDM2.Close()

	// Non-participant User 3 attempting to view DM page -> should redirect to /chat
	respViewDM3, err := ts.Client.R().
		SetHeader("Cookie", cookie3).
		Get(fmt.Sprintf("/chat/dm/%d", dmChannel.Id))
	if err != nil {
		t.Fatalf("non-participant DM page view failed: %v", err)
	}
	defer respViewDM3.Close()
	if respViewDM3.StatusCode() == http.StatusOK && strings.Contains(respViewDM3.String(), "DM User Two") {
		t.Errorf("expected non-participant user3 to be denied access to DM page, but got 200 OK")
	}

	// 4. Send DM Message (User 1 -> User 2)
	respSendMsg, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"content": "Hello in secret DM!",
		}).
		Post(fmt.Sprintf("/groups/dm/channels/%d/messages/create", dmChannel.Id))
	if err != nil || respSendMsg.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for sending DM message, got status: %v (err: %v)", respSendMsg.StatusCode(), err)
	}
	respSendMsg.Close()

	// Verify notification for User 2
	notifsUser2, err := model.GetUserNotifications(user2.Id, 10)
	if err != nil || len(notifsUser2) == 0 {
		t.Fatalf("expected notification for user2 on DM message, got: %v", notifsUser2)
	}
	if notifsUser2[0].Type != "dm" || !strings.Contains(notifsUser2[0].Content, "Hello in secret DM!") {
		t.Errorf("expected DM notification content, got: %v", notifsUser2[0])
	}

	// Non-participant User 3 attempting to send message in DM channel -> should be forbidden
	respNonPartMsg, err := ts.Client.R().
		SetHeader("Cookie", cookie3).
		SetFormDataWithMap(map[string]string{
			"content": "Unwanted intruder message",
		}).
		Post(fmt.Sprintf("/groups/dm/channels/%d/messages/create", dmChannel.Id))
	if err != nil {
		t.Fatalf("non-participant DM post failed: %v", err)
	}
	defer respNonPartMsg.Close()
	if respNonPartMsg.StatusCode() == http.StatusOK {
		t.Errorf("expected non-participant DM message post to fail, got 200 OK")
	}

	// 5. GET /groups/dm/channels/:channel_id/messages/page/1
	respDMPage, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/dm/channels/%d/messages/page/1", dmChannel.Id))
	if err != nil || respDMPage.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for DM message pagination, got status: %v", respDMPage.StatusCode())
	}
	defer respDMPage.Close()
	if !strings.Contains(respDMPage.String(), "Hello in secret DM!") {
		t.Errorf("expected DM message page body to contain message content, got: %s", respDMPage.String())
	}

	_ = user3
}

func TestChannelStarringAndEditing(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Star Owner", "star_owner", "star_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Star Group", "star_group")

	ch := model.NewChannel("star_ch", "Star Channel", "Star description", model.ChannelTypeText)
	if err := ch.Insert(); err != nil {
		t.Fatalf("failed to insert channel: %v", err)
	}
	_ = model.AddGroupChannel(group.Id, ch.Id, 0)

	// 1. Star Channel
	respStar, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/channels/%d/star", group.Id, ch.Id))
	if err != nil || respStar.StatusCode() != http.StatusOK {
		t.Fatalf("failed to star channel: %v", err)
	}
	defer respStar.Close()

	if !model.IsChannelStarred(owner.Id, ch.Id) {
		t.Errorf("expected channel %d to be starred in DB", ch.Id)
	}
	if !strings.Contains(respStar.String(), "Unstar channel") {
		t.Errorf("expected response button to contain 'Unstar channel', got: %s", respStar.String())
	}

	// 2. Unstar Channel
	respUnstar, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/channels/%d/star", group.Id, ch.Id))
	if err != nil || respUnstar.StatusCode() != http.StatusOK {
		t.Fatalf("failed to unstar channel: %v", err)
	}
	defer respUnstar.Close()

	if model.IsChannelStarred(owner.Id, ch.Id) {
		t.Errorf("expected channel %d to no longer be starred in DB", ch.Id)
	}
	if !strings.Contains(respUnstar.String(), "Star channel") {
		t.Errorf("expected response button to contain 'Star channel', got: %s", respUnstar.String())
	}

	// 3. Edit Channel
	respEdit, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"name":        "star_ch_mod",
			"moniker":     "Star Channel Modified",
			"description": "Updated description",
		}).
		Put(fmt.Sprintf("/groups/%d/channels/%d", group.Id, ch.Id))
	if err != nil || (respEdit.StatusCode() != http.StatusOK && respEdit.StatusCode() != http.StatusSeeOther) {
		t.Fatalf("failed to edit channel: %v", err)
	}
	respEdit.Close()

	updatedCh, err := model.NewChannelFromId(ch.Id)
	if err != nil || updatedCh.Moniker != "Star Channel Modified" || updatedCh.Description != "Updated description" {
		t.Errorf("expected updated channel moniker 'Star Channel Modified', got: %v", updatedCh)
	}
}
